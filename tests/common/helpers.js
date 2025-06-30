
import { DELETE, GET, POST } from './request'

async function deleteAllNamespaces () {
	const listResponse = await GET(`/api/v2/namespaces`)

	if (listResponse.statusCode !== 200)
		throw Error(`none ok namespaces list statusCode(${ listResponse.statusCode })`)

	for (const namespace of listResponse.body.data) {
		const response = await DELETE(`/api/v2/namespaces/${ namespace.name }`)

		if (response.statusCode !== 200)
			throw Error(`none ok namespace(${ namespace.name }) delete statusCode(${ response.statusCode })`)
	}
}

async function itShouldCreateNamespace (it, expect, ns) {
	it(`should create a new namespace ${ ns }`, async () => {
		const res = await POST(`/api/v2/namespaces`)
			.send({ name: ns })
		expect(res.statusCode).toEqual(200)
	})
}

function sleep (ms) {
	return new Promise(resolve => setTimeout(resolve, ms))
}

export default {
	deleteAllNamespaces,
	itShouldCreateNamespace,
	sleep,
}
