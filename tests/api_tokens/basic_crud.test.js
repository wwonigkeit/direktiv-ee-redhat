import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import regex from '../common/regex'
import { DELETE, GET, POST } from '../common/request'

const namespace = basename(__filename)

describe('Test api_tokens get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should create a new api_token foo1`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/api_tokens`)
			.send(makeDummyAPIToken('foo1'))
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			apiToken: expectDummyAPIToken('foo1'),
			secret: expect.stringMatching(regex.uuidRegex),
		},
		)
	})

	it(`should create a new api_token foo2`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/api_tokens`)
			.send(makeDummyAPIToken('foo2'))
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new api_token foo1`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/api_tokens/foo1`)

		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toEqual(expectDummyAPIToken('foo1'))
	})

	it(`should get the new api_token foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/api_tokens/foo2`)

		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual(expectDummyAPIToken('foo2'))
	})

	it(`should list foo1 and foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/api_tokens`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyAPIToken('foo1'), expectDummyAPIToken('foo2') ],
		})
	})

	it(`should delete foo1`, async () => {
		const res = await DELETE(`/api/v2/namespaces/${ namespace }/api_tokens/foo1`)

		expect(res.statusCode).toEqual(200)
	})

	it(`should list foo2`, async () => {
		const res = await GET(`/api/v2/namespaces/${ namespace }/api_tokens`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ expectDummyAPIToken('foo2') ],
		})
	})

	it(`should create a new api_token with no perms`, async () => {
		const res = await POST(`/api/v2/namespaces/${ namespace }/api_tokens`)
			.send({
				name: 'foo',
				description: 'description',
				permissions: null,
				duration: 'P1DT2H30M',
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			apiToken: {
				name: 'foo',
				description: 'description',
				permissions: null,
				prefix: expect.anything(),
				isExpired: false,
				expiredAt: expect.stringMatching(regex.timestampRegex),
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			},
			secret: expect.stringMatching(regex.uuidRegex),
		},
		)
	})
})

function makeDummyAPIToken (name) {
	return {
		name,
		description: name + ' description',
		permissions: [ {
			topic: 'secrets',
			method: 'read',
		}, {
			topic: 'variables',
			method: 'manage',
		} ],
		duration: 'P1DT2H30M',
	}
}

function expectDummyAPIToken (name) {
	return {
		name,
		description: name + ' description',
		prefix: expect.anything(),
		permissions: [ {
			topic: 'secrets',
			method: 'read',
		}, {
			topic: 'variables',
			method: 'manage',
		} ],
		isExpired: false,
		expiredAt: expect.stringMatching(regex.timestampRegex),
		createdAt: expect.stringMatching(regex.timestampRegex),
		updatedAt: expect.stringMatching(regex.timestampRegex),
	}
}
