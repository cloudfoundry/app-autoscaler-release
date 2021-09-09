require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'scheduler' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('scheduler') }
  let(:template) { job.template('config/application.properties') }
  let(:properties) do
    YAML.safe_load(%(
      autoscaler:
        scheduler_db:
          address: 10.11.137.101
          databases:
          - name: foo
            password: default
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
        policy_db:
          address: 10.11.137.101
          databases:
          - name: foo
            password: default
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
    ))
  end

  context 'config/application.properties' do

    it 'does not set username nor password if not configured' do
      properties['autoscaler'].merge!(
        'scheduler' => {
          'health' => {
            'port' => 1234,
          }
        }
      )
      rendered_template = template.render(properties)

      expect(rendered_template).to include('scheduler.healthserver.port=1234')
      expect(rendered_template).to include('scheduler.healthserver.basicAuthEnabled=false')
      expect(rendered_template).to include('scheduler.healthserver.username=')
      expect(rendered_template).to include('scheduler.healthserver.password=')
    end

    it 'check scheduler username and password' do
      properties['autoscaler'].merge!(
        'scheduler' => {
          'health' => {
            'port' => 1234,
            'basicAuthEnabled' => 'true',
            'username' => 'test-user',
            'password' => 'test-user-password'
          }
        }
      )
      rendered_template = template.render(properties)

      expect(rendered_template).to include('scheduler.healthserver.port=1234')
      expect(rendered_template).to include('scheduler.healthserver.basicAuthEnabled=true')
      expect(rendered_template).to include('scheduler.healthserver.username=test-user')
      expect(rendered_template).to include('scheduler.healthserver.password=test-user-password')

    end
  end
end
