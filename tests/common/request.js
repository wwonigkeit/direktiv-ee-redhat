import request from 'supertest'

import config from './config'

const POST = function (path) {
	 return request(config.getDirektivHost())
		.post(path)
		.set('Direktiv-Api-Key', 'password')
}

const GET = function (path) {
	return request(config.getDirektivHost())
		.get(path)
		.set('Direktiv-Api-Key', 'password')
}

const DELETE = function (path) {
	return request(config.getDirektivHost())
		.delete(path)
		.set('Direktiv-Api-Key', 'password')
}

const PUT = function (path) {
	return request(config.getDirektivHost())
		.put(path)
		.set('Direktiv-Api-Key', 'password')
}

export { DELETE, GET, POST, PUT }
