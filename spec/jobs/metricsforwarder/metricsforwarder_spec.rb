require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'metricsforwarder' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('metricsforwarder') }
  let(:template) { job.template('config/metricsforwarder.yml') }
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
        cf:
          api: https://api.cf.domain
          auth_endpoint: https://login.cf.domain
          client_id: client_id
          secret: uaa_secret
          uaa_api: https://login.cf.domain/uaa
          grant_type: ALLOW_ALL
    ))
  end

  context 'config/metricsforwarder.yml' do

    it 'does not set username nor password if not configured' do
      properties['autoscaler'].merge!(
        'metricsforwarder' => {
          'health' => {
            'port' => 1234,
          }
        }
      )
      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template['health']).
        to include(
             { 'port' => 1234 }
           )
    end

    it 'check metricsforwarder basic auth username and password' do
      properties['autoscaler'].merge!(
        'metricsforwarder' => {
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

    it 'has no instance identity configured by default' do
      rendered_template = YAML.safe_load(template.render(properties))
      expect(rendered_template).
        to_not include( 'metrics_forwarder' )
    end

    it 'has instance identity configured' do
      properties['autoscaler'].merge!(
        'metricsforwarder' => {
          'metricshandler' => {
            'tls' => {
              'ca_cert' => '--- THIS IS A CERTIFICATE ---',
            }
          }
        }
      )

      rendered_template = YAML.safe_load(template.render(properties))
      expect(rendered_template).
        to include(
             {
               "metrics_forwarder" => {
                 "tls"=> {
                   "ca_file" => "/var/vcap/jobs/metricsforwarder/config/certs/instance_identity_ca.crt"
                 }
               }
             }
           )
    end
  end
end

