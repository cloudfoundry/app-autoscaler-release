const request = require('supertest')
// const nock = require('nock')
const app = require('./app')

describe('app', () => {
    describe('get /fast', () => {
        it('return fast response', async () => {
            const res = await request(app).get('/fast')
            expect(res.statusCode).toEqual(200)
            expect(res.text).toEqual('dummy application with fast response')
        })
    })


    describe('when autoscaler service is binded to the app', function () {
        describe('get /custom-metrics/mtls/:type/:value', () => {
            xit('writes metric successfully', async () => {
                process.env.VCAP_APPLICATION = JSON.stringify({
                    "application_id": "some_guid",
                })

                process.env.VCAP_SERVICES = JSON.stringify({
                    "autoscaler": [
                        {
                            "credentials": {
                                "custom_metrics": {
                                    "mtls_url": "https://autoscaler-metrics-mtls.cf.test.com",
                                    "url": "https://autoscaler-metrics.cf.test.com",
                                    "username": "some_user",
                                    "password": "some_password"
                                }
                            },
                        }
                    ]
                })

                process.env.CF_INSTANCE_KEY = 'path_to_key'
                process.env.CF_INSTANCE_CERT = 'path_to_cert'
            })
        })
    })
})
