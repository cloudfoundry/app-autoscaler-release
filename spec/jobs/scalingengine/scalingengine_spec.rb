require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'scalingengine' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('scalingengine') }
  let(:template) { job.template('config/scalingengine.yml') }
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
        instancemetrics_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          tls.ca: aa
          roles:
          - name: foo
            password: default
            tag: default
        appmetrics_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            tag: default
        scalingengine_db:
          address: 10.11.137.105
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 6543
          roles:
          - name: foo
            password: default
            tag: default
        lock_db:
          address: 10.11.137.105
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 6543
          roles:
          - name: foo
            tag: default
        scheduler_db:
          address: 10.11.137.105
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 6543
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

  context 'config/scalingengine.yml' do

    it 'does not set username nor password if not configured' do
      properties['autoscaler'].merge!(
        'scalingengine' => {
          'health' => {
            'port' => 1234
          }
        }
      )
      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template['health']).
        to include(
             { 'port' => 1234 }
           )
    end

    it 'check scalingengine basic auth username and password' do
      properties['autoscaler'].merge!(
        'scalingengine' => {
          'health' => {
            'port' => 1234,
            'username' => 'test-user',
            'password' => 'test-user-password'
          }
        }
      )
      rendered_template = YAML.safe_load(template.render(properties))

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

