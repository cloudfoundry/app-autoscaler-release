require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'metricsgateway' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('metricsgateway') }
  let(:template) { job.template('config/metricsgateway.yml') }
  let(:properties) do
    YAML.safe_load(%(
      autoscaler:
        policy_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
        sbss_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
        cf:
          api: https://api.cf.domain
          auth_endpoint: https://login.cf.domain
          client_id: client_id
          secret: uaa_secret
          uaa_api: https://login.cf.domain/uaa
          grant_type: ALLOW_ALL
    ))
  end

  context 'config/metricsgateway.yml' do

    it 'does not set username nor password if not configured' do
      properties['autoscaler'].merge!(
        'metricsgateway' => {
          'health' => {
            'port' => 1234
          },
          'nozzle' => {
            'rlp_addr' => 'localhost'
          }
        }
      )
      links = [
        Bosh::Template::Test::Link.new(
          name: 'metricsserver'
        )
      ]
      rendered_template = YAML.safe_load(template.render(properties, consumes: links))

      expect(rendered_template['health']).
        to include(
             { 'port' => 1234,
             }
           )
    end

    it 'check metricsgateway basic auth username and password' do
      properties['autoscaler'].merge!(
        'metricsgateway' => {
          'health' => {
            'port' => 1234,
            'username' => 'test-user',
            'password' => 'test-user-password'
          },
          'nozzle' => {
            'rlp_addr' => 'localhost'
          }
        }
      )
      links = [
        Bosh::Template::Test::Link.new(
          name: 'metricsserver'
        )
      ]
      rendered_template = YAML.safe_load(template.render(properties, consumes: links))

      expect(rendered_template['health']).
        to include(
             { 'port' => 1234,
               'username' => 'test-user',
               'password' => 'test-user-password'
             }
           )
    end
  end
end

