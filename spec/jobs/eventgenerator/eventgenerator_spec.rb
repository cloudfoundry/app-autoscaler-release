require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'eventgenerator' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('eventgenerator') }
  let(:template) { job.template('config/eventgenerator.yml') }
  let(:properties) do
    YAML.safe_load(%(
      autoscaler:
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
        appmetrics_db:
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

  context 'config/eventgenerator.yml' do

    it 'does not set username nor password if not configured' do
      properties['autoscaler'].merge!(
        'eventgenerator' => {
          'health' => {
            'port' => 1234,
          }
        }
      )
      links = [
        Bosh::Template::Test::Link.new(
          name: 'eventgenerator',
          properties: {}
        )
      ]
      rendered_template = YAML.safe_load(template.render(properties, consumes: links))
      expect(rendered_template['health']).
        to include(
             { 'port' => 1234 }
           )
    end

    it 'check eventgenerator username and password' do
      properties['autoscaler'].merge!(
        'eventgenerator' => {
          'health' => {
            'port' => 1234,
            'username' => 'test-user',
            'password' => 'test-user-password'
          }
        }
      )
      links = [
        Bosh::Template::Test::Link.new(
          name: 'eventgenerator',
          properties: {}
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
