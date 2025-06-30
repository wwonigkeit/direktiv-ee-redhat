import { beforeAll, describe, expect, it } from '@jest/globals'
import request from 'supertest'

import config from '../common/config'
import helpers from '../common/helpers'
import { GET, POST } from '../common/request'

describe('test api tokens authorization logic', () => {
	const allCombinations = [];

	[ 'p1', 'p2', 'p3' ].forEach(pr => {
		[ 'GET', 'POST', 'PUT', 'manage', 'read' ].forEach(method => {
			[ 'secrets', 'variables', 'instances' ].forEach(topic => {
				allCombinations.push({ pr, method, topic })
			})
		})
	})

	const tokens = {}

	beforeAll(async () => {
		await helpers.deleteAllNamespaces()
		for (const pr of [ 'p1', 'p2', 'p3' ]) {
			const res = await POST('/api/v2/namespaces').send({
				name: pr,
			})
			expect(res.statusCode).toEqual(200)
		}

		for (let i = 0; i < allCombinations.length; i++) {
			const { pr, method, topic } = allCombinations[i]

			const tokenTitle = `pr_${ pr }_method_${ method }_topic_${ topic }`

			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces/${ pr }/api_tokens`)
				.set('Direktiv-Api-Key', 'password')
				.send({
					name: tokenTitle,
					description: 'des1',
					duration: 'PT1M',
					permissions: [ {
						method,
						topic,
					} ],
				})
			expect(res.statusCode).toEqual(200)
			tokens[tokenTitle] = res.body.data.secret
		}
	})

	for (let i = 0; i < allCombinations.length; i++) {
		const req = allCombinations[i]

		if (req.method === 'manage' || req.method === 'read') continue

		const reqTitle = `pr_${ req.pr }_method_${ req.method }_topic_${ req.topic }`

		for (let j = 0; j < allCombinations.length; j++) {
			const sample = allCombinations[j]

			const tokenTitle = `pr_${ sample.pr }_method_${ sample.method }_topic_${ sample.topic }`

			let itShouldAccess = true
			if (req.pr !== sample.pr) itShouldAccess = false

			if (req.method === 'GET' && sample.method === 'read') {
			} else if (req.method !== sample.method && sample.method !== 'manage')
				itShouldAccess = false

			if (req.topic !== sample.topic)
				itShouldAccess = false

			it(`should ${ itShouldAccess === true ? '' : 'NOT' } access request(${ reqTitle }) endpoint with token(${ tokenTitle })`, async () => {
				const res = await request(config.getDirektivHost())
					[req.method.toLowerCase()](`/api/v2/namespaces/${ req.pr }/${ req.topic }/something`)
					.set('Direktiv-Api-Key', 'password')
					.set('Direktiv-Api-Token', tokens[tokenTitle])
					.send()

				if (itShouldAccess)
					expect([ 200, 400, 404, 405 ]).toContain(res.statusCode)

				else expect(res.statusCode).toEqual(403)
			})
		}
	}
})
