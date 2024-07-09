require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"
require_relative "../utils"

describe "golangapiserver" do
  context "apiserver.yml.erb" do
    let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
    let(:job) { release.job("golangapiserver") }
    let(:template) { job.template("config/apiserver.yml") }
    let(:properties) { YAML.safe_load(fixture("apiserver.yml").read) }
    let(:rendered_template) { YAML.safe_load(template.render(properties)) }

    context "handles broker credentials" do
      it "writes service_broker_usernames" do
        properties["autoscaler"]["apiserver"]["broker"]["broker_credentials"] = [
          {"broker_username" => "fake_b_user_1",
           "broker_password" => "fake_b_password_1"},
          {"broker_username" => "fake_b_user_2",
           "broker_password" => "fake_b_password_2"}
        ]

        rendered_template = YAML.safe_load(template.render(properties))

        expect(rendered_template["broker_credentials"]).to include(
          {"broker_username" => "fake_b_user_1",
           "broker_password" => "fake_b_password_1"},
          {"broker_username" => "fake_b_user_2",
           "broker_password" => "fake_b_password_2"}
        )
      end

      it "writes deprecated service_broker_usernames" do
        properties["autoscaler"]["apiserver"]["broker"]["broker_credentials"] = nil
        properties["autoscaler"]["apiserver"]["broker"].merge!(
          "username" => "deprecated_username",
          "password" => "deprecated_password"
        )

        rendered_template = YAML.safe_load(template.render(properties))

        expect(rendered_template["broker_credentials"]).to include(
          {"broker_username" => "deprecated_username",
           "broker_password" => "deprecated_password"}
        )
      end

      it "favour list of credentials over deprecated values" do
        properties["autoscaler"]["apiserver"]["broker"].merge!(
          "broker_credentials" => [
            {"broker_username" => "fake_b_user_1",
             "broker_password" => "fake_b_password_1"},
            {"broker_username" => "fake_b_user_2",
             "broker_password" => "fake_b_password_2"}
          ],
          "username" => "deprecated_username",
          "password" => "deprecated_password"
        )

        rendered_template = YAML.safe_load(template.render(properties))

        expect(rendered_template["broker_credentials"]).to include(
          {"broker_username" => "fake_b_user_1",
           "broker_password" => "fake_b_password_1"},
          {"broker_username" => "fake_b_user_2",
           "broker_password" => "fake_b_password_2"}
        )
      end

      context "has broker credentials set up" do
        before(:each) do
          properties["autoscaler"]["apiserver"]["broker"]["broker_credentials"] = [
            {"broker_username" => "fake_b_user_1",
             "broker_password" => "fake_b_password_1"},
            {"broker_username" => "fake_b_user_2",
             "broker_password" => "fake_b_password_2"}
          ]
        end

        it "by default TLS is not configured" do
          rendered_template = YAML.safe_load(template.render(properties))

          expect(rendered_template["broker_server"]["tls"]).to be_nil
        end

        it "TLS can be enabled" do
          properties["autoscaler"]["apiserver"]["broker"]["server"].merge!({
            "ca_cert" => "SOME_CA",
            "server_cert" => "SOME_CERT",
            "server_key" => "SOME_KEY"
          })

          rendered_template = YAML.safe_load(template.render(properties))

          expect(rendered_template["broker_server"]["tls"]).not_to be_nil
          expect(rendered_template["broker_server"]["tls"]).to include({
            "key_file" => "/var/vcap/jobs/golangapiserver/config/certs/brokerserver/server.key",
            "ca_file" => "/var/vcap/jobs/golangapiserver/config/certs/brokerserver/ca.crt",
            "cert_file" => "/var/vcap/jobs/golangapiserver/config/certs/brokerserver/server.crt"
          })
        end
      end
    end

    context "plan_check" do
      it "by default plan checks are disabled" do
        expect(rendered_template["plan_check"]).to be_nil
      end

      it "plan checks can be enabled" do
        properties["autoscaler"]["apiserver"]["broker"]["plan_check"] = {
          "plan_definitions" => {
            "Some-example-uuid-ONE" => {"planCheckEnabled" => true, "schedules_count" => 2, "scaling_rules_count" => 4},
            "Some-example-uuid-TWO" => {"planCheckEnabled" => true, "schedules_count" => 10, "scaling_rules_count" => 10}
          }
        }

        rendered_template = YAML.safe_load(template.render(properties))

        expect(rendered_template["plan_check"]).to include(
          {"plan_definitions" => {
            "Some-example-uuid-ONE" => {"planCheckEnabled" => true, "scaling_rules_count" => 4, "schedules_count" => 2},
            "Some-example-uuid-TWO" => {"planCheckEnabled" => true, "scaling_rules_count" => 10, "schedules_count" => 10}
          }}
        )
      end
    end

    context "cred_helper_impl" do
      it "has a cred helper impl by default" do
        expect(rendered_template).to include(
          {
            "cred_helper_impl" => "default"
          }
        )
      end

      it "has a cred helper impl configured for stored procedures" do
        properties["autoscaler"]["apiserver"]["cred_helper"] = {
          "impl" => "stored_procedure",
          "stored_procedure_config" => {
            "schema_name" => "SCHEMA",
            "create_binding_credential_procedure_name" => "CREATE_BINDING_CREDENTIAL",
            "drop_binding_credential_procedure_name" => "DROP_BINDING_CREDENTIAL",
            "drop_all_binding_credential_procedure_name" => "DROP_ALL_BINDING_CREDENTIALS",
            "validate_binding_credential_procedure_name" => "VALIDATE_BINDING_CREDENTIALS"
          }
        }

        rendered_template = YAML.safe_load(template.render(properties))
        expect(rendered_template).to include(
          {
            "cred_helper_impl" => "stored_procedure",
            "stored_procedure_binding_credential_config" => {
              "schema_name" => "SCHEMA",
              "create_binding_credential_procedure_name" => "CREATE_BINDING_CREDENTIAL",
              "drop_binding_credential_procedure_name" => "DROP_BINDING_CREDENTIAL",
              "drop_all_binding_credential_procedure_name" => "DROP_ALL_BINDING_CREDENTIALS",
              "validate_binding_credential_procedure_name" => "VALIDATE_BINDING_CREDENTIALS"
            }
          }
        )
      end
    end

    context "storedprocedure_db" do
      it "selects db role with storedproceduredb tag by default" do
        rendered_template["db"]["storedprocedure_db"]["url"].tap do |url|
          expect(url).to include("stored_procedure_username")
          expect(url).to include("store_procedure_db")
        end
      end
    end

    context "uses tls" do
      context "binding_db" do
        before do
          properties["autoscaler"]["apiserver"]["use_buildin_mode"] = false
          properties["autoscaler"]["apiserver"]["broker"]["username"] = "foouser"
          properties["autoscaler"]["apiserver"]["broker"]["password"] = "foopw"
        end

        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["binding_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "binding_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["binding_db"]["tls"] = nil
          rendered_template["db"]["binding_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "binding_db")
          end
        end
      end

      context "policy_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["policy_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "policy_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["policy_db"]["tls"] = nil
          rendered_template["db"]["policy_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "policy_db")
          end
        end
      end

      context "storedprocedure_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["storedprocedure_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "storedprocedure_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["storedprocedure_db"]["tls"] = nil
          rendered_template["db"]["storedprocedure_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "storedprocedure_db")
          end
        end
      end
    end
  end
end
