const path = require('path');
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const jwt = require('jsonwebtoken');

const SECRET_KEY = 'supersecret';
const users = { user1: 'password123' };

const packageDefinition = protoLoader.loadSync(
  path.join(__dirname, 'auth', 'auth.proto'),
  {
	includeDirs: [
	  path.join(__dirname, 'googleapis'), // Путь к скачанному googleapis
	],
	keepCase: true,
  }
);
const authProto = grpc.loadPackageDefinition(packageDefinition).auth;

function login(call, callback) {
  const { username, password } = call.request;
  if (users[username] && users[username] === password) {
    const token = jwt.sign({ username }, SECRET_KEY, { expiresIn: '1h' });
    callback(null, { access_token: token });
  } else {
    callback({ code: grpc.status.UNAUTHENTICATED, message: 'Invalid credentials' });
  }
}

function validateToken(call, callback) {
  const { access_token } = call.request;
  jwt.verify(access_token, SECRET_KEY, (err, decoded) => {
    if (err) {
      callback(null, { valid: false });
    } else {
      callback(null, { username: decoded.username, valid: true });
    }
  });
}

function main() {
  const server = new grpc.Server();
  server.addService(authProto.AuthService.service, { login, validateToken });
  server.bindAsync('0.0.0.0:50052', grpc.ServerCredentials.createInsecure(), () => {
    console.log("AuthService running on port 50052");
  });
}

main();
