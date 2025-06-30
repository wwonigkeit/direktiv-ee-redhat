import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import regex from '../common/regex'
import { DELETE, GET, POST, PUT } from '../common/request'

const namespace = basename(__filename)

describe('Test roles get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new role foo1`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/roles`)
			.send(makeDummyRole('foo1'))
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(
			expectDummyRole('foo1'))
	})

	it(`should create a new role foo2`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/roles`)
			.send(makeDummyRole('foo2'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new role foo1`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/roles/foo1`)

		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummyRole('foo1'))
	})

	it(`should get the new role foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/roles/foo2`)

		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummyRole('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/roles`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyRole('foo1'), expectDummyRole('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await DELETE(`/api/v2/namespaces/${ namespace }/roles/foo1`)

		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/roles`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyRole('foo2') ],
		})
	})

	it(`should update foo2`, async () => {
		const res = await PUT(`/api/v2/namespaces/${ namespace }/roles/foo2`)
			.send(makeDummyRole('foo3'))
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(
			expectDummyRole('foo3'))
	})

	it(`should create a new role foo4 without oidc groups`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/roles`)
			.send({
				name: 'foo4',
				description: 'description',
				oidcGroups: null,
				permissions: [ {
					topic: 'secrets',
					method: 'read',
				}, {
					topic: 'variables',
					method: 'manage',
				} ],
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			name: 'foo4',
			description: 'description',
			oidcGroups: null,
			permissions: [ {
				topic: 'secrets',
				method: 'read',
			}, {
				topic: 'variables',
				method: 'manage',
			} ],
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
	it(`should create a new role foo5 without perms`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/roles`)
			.send({
				name: 'foo5',
				description: 'description',
				oidcGroups: null,
				permissions: [],
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			name: 'foo5',
			description: 'description',
			oidcGroups: null,
			permissions: [],
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		})
	})
})

function makeDummyRole (name) {
	return {
		name,
		description: name + ' description',
		oidcGroups: [ name + '_g1', name + '_g2' ],
		permissions: [ {
			topic: 'secrets',
			method: 'read',
		}, {
			topic: 'variables',
			method: 'manage',
		} ],
	}
}

function expectDummyRole (name) {
	return {
		name,
		description: name + ' description',
		oidcGroups: [ name + '_g1', name + '_g2' ],
		permissions: [ {
			topic: 'secrets',
			method: 'read',
		}, {
			topic: 'variables',
			method: 'manage',
		} ],
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
