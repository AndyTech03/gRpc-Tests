const path = require('path');
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const packageDefinition = protoLoader.loadSync(
	path.join(__dirname, 'hello', 'hello.proto'),
	{ 
		includeDirs: [
			path.join(__dirname, 'googleapis'), // Путь к скачанному googleapis
		],
		keepCase: true 
	}
);
const helloProto = grpc.loadPackageDefinition(packageDefinition).hello;

function sayHello(call, callback) {
  const { language } = call.request || 'en';
  const username = call.metadata.get('username')[0] || 'Guest';
  const authErr = call.metadata.get('err')[0];
  const authErrCode = call.metadata.get('errCode')[0];

  if (authErr) {
    console.log({ authErrCode, authErr })
    return callback({ code: Number(authErrCode), message: authErr });
  }

  const greetings = {
    en: `Hello, ${username}!`,
    ru: `Привет, ${username}!`
  };
  
  const reply = { message: greetings[language] || greetings.en };
  callback(null, reply);
}

function main() {
  const server = new grpc.Server();
  server.addService(helloProto.HelloService.service, { SayHello: sayHello });
  server.bindAsync('0.0.0.0:50051', grpc.ServerCredentials.createInsecure(), () => {
    console.log("HelloService running on port 50051");
  });
}

main();
