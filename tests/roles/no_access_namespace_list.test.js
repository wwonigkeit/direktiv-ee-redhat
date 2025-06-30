import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import { GET, POST } from '../common/request'

const namespace = basename(__filename)

describe('Test role should not access GET /namespaces', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace + '1')
	helpers.itShouldCreateNamespace(it, expect, namespace + '2')

	it(`should create role foo1`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace + '1' }/roles`)
			.send({
				name: 'foo1',
				description: 'description',
				oidcGroups: [ 'g1' ],
				permissions: [ {
					topic: 'namespaces',
					method: 'read',
				} ],
			})
		expect(res.statusCode).toEqual(200)
	})

	it(`should create role foo2`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace + '2' }/roles`)
			.send({
				name: 'foo2',
				description: 'description',
				oidcGroups: [ 'g2' ],
				permissions: [ {
					topic: 'namespaces',
					method: 'read',
				} ],
			})
		expect(res.statusCode).toEqual(200)
	})

	it(`should get namespace1 with g1`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace + '1' }`)
			.set('X-Oidc-Groups', 'g1')
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.name).toEqual(namespace + '1')
	})
	it(`should get namespace2 with g2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace + '2' }`)
			.set('X-Oidc-Groups', 'g2')
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.name).toEqual(namespace + '2')
	})

	it(`should get namespace1 with g1`, async () => {
		const res = await GET(`/api/v2/namespaces`)
			.set('X-Oidc-Groups', 'g1')
		expect(res.statusCode).toEqual(200)
		const gotNamespaces = res.body.data.map(i => i.name)
		expect(gotNamespaces).toEqual([ namespace + '1' ])
	})

	it(`should get namespace2 with g2`, async () => {
		const res = await GET(`/api/v2/namespaces`)
			.set('X-Oidc-Groups', 'g2')
		expect(res.statusCode).toEqual(200)
		const gotNamespaces = res.body.data.map(i => i.name)
		expect(gotNamespaces).toEqual([ namespace + '2' ])
	})

	it(`should get all namespaces with full access`, async () => {
		const res = await GET(`/api/v2/namespaces`)
		expect(res.statusCode).toEqual(200)
		const gotNamespaces = res.body.data.map(i => i.name)
		expect(gotNamespaces).toEqual([ namespace + '1', namespace + '2' ])
	})

	it(`should not allowed to create namespaces`, async () => {
		const res = await POST(`/api/v2/namespaces`)
			.set('X-Oidc-Groups', 'g1')
			.send({})
		expect(res.statusCode).toEqual(403)
	})
})
