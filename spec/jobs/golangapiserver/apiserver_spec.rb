require 'rspec'
require 'json'
require 'bosh/template/test'
require 'yaml'

describe 'golangapiserver' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../../..')) }
  let(:job) { release.job('golangapiserver') }
  let(:template) { job.template('config/apiserver.yml') }
  let(:properties) do
    YAML.safe_load(%(
      autoscaler:
        binding_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            tag: default
        policy_db:
          address: 10.11.137.101
          databases:
          - name: foo
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
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
        apiserver:
          broker:
            server:
              dashboard_redirect_uri: https://application-autoscaler-dashboard.cf.domain
            plan_check: --
    ))
  end

  context 'config/apiserver.yml' do
    context 'apiserver does not use buildin mode' do
      before(:each) do
        properties['autoscaler']['apiserver'].merge!(
          'use_buildin_mode' => false,
        )
      end

      it 'writes service_broker_usernames' do
        properties['autoscaler']['apiserver']['broker'].merge!(
          'broker_credentials' => [
            { 'broker_username' => 'fake_b_user_1',
              'broker_password' => 'fake_b_password_1' },
            { 'broker_username' => 'fake_b_user_2',
              'broker_password' => 'fake_b_password_2' },
          ],
        )

        rendered_template = YAML.safe_load(template.render(properties))

        expect(rendered_template['broker_credentials']).to include(
          { 'broker_username' => 'fake_b_user_1',
            'broker_password' => 'fake_b_password_1' },
          { 'broker_username' => 'fake_b_user_2',
            'broker_password' => 'fake_b_password_2' },
        )
      end
    end
  end
end
