import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import { GET, POST } from '../common/request'

const namespace = basename(__filename)

describe('Test api_tokens should not access GET /namespaces', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace + '1')
	helpers.itShouldCreateNamespace(it, expect, namespace + '2')

	let token

	it(`should create a new api_token foo1`, async () => {
		token = await POST(`/api/v2/namespaces/${ namespace + '1' }/api_tokens`)
			.send({
				name: 'foo',
				description: 'description',
				permissions: [ {
					topic: 'namespaces',
					method: 'manage',
				} ],
				duration: 'P1DT2H30M',
			})
		expect(token.statusCode).toEqual(200)
	})

	it(`should get namespace1`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace + '1' }`)
			.set('Direktiv-Api-Token', token.body.data.secret)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.name).toEqual(namespace + '1')
	})

	it(`should only get namespace1 on /namespaces`, async () => {
		const res = await GET(`/api/v2/namespaces`)
			.set('Direktiv-Api-Token', token.body.data.secret)
		expect(res.statusCode).toEqual(200)
		const gotNamespaces = res.body.data.map(i => i.name)
		expect(gotNamespaces).toEqual([ namespace + '1' ])
	})

	it(`should get all namespaces with full access`, async () => {
		const res = await GET(`/api/v2/namespaces`)
		expect(res.statusCode).toEqual(200)
		const gotNamespaces = res.body.data.map(i => i.name)
		expect(gotNamespaces).toEqual([ namespace + '1', namespace + '2' ])
	})

	it(`should not allowed to create namespaces`, async () => {
		const res = await POST(`/api/v2/namespaces`)
			.set('Direktiv-Api-Token', token.body.data.secret)
			.send({})
		expect(res.statusCode).toEqual(403)
	})
})
