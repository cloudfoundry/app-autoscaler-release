require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"
require_relative "../utils"

describe "metricsforwarder" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("metricsforwarder") }
  let(:template) { job.template("config/metricsforwarder.yml") }
  let(:properties) { YAML.safe_load(fixture("metricsforwarder.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/metricsforwarder.yml" do
    it "supports syslog forwarding" do
      properties["autoscaler"]["metricsforwarder"] = {
        "syslog" => {
          "server_address" => "syslog-server"
        }
      }

      expect(rendered_template).to include(
        {
          "syslog" => {
            "server_address" => "syslog-server",
            "port" => 6067,
            "tls" => {
              "key_file" => "/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/client.key",
              "cert_file" => "/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/client.crt",
              "ca_file" => "/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/ca.crt"
            }
          }
        }
      )
    end

    it "does not set username nor password if not configured" do
      properties["autoscaler"]["metricsforwarder"] = {
        "health" => { }
      }

      expect(rendered_template["health"])
        .to include({})
    end

    it "check metricsforwarder basic auth username and password" do
      properties["autoscaler"]["metricsforwarder"] = {
        "health" => {
          "username" => "test-user",
          "password" => "test-user-password"
        }
      }

      expect(rendered_template["health"])
        .to include(
          {
           "basic_auth" => {
           "username" => "test-user",
           "password" => "test-user-password"} }
        )
    end

    it "has a cred helper impl by default" do
      expect(rendered_template).to include({
        "cred_helper_impl" => "default"
      })
    end

    it "has a cred helper impl configured for stored procedures" do
      properties["autoscaler"]["metricsforwarder"] = {
        "cred_helper" => {
          "impl" => "stored_procedure",
          "stored_procedure_config" => {
            "schema_name" => "SCHEMA",
            "create_binding_credential_procedure_name" => "CREATE_BINDING_CREDENTIAL",
            "drop_binding_credential_procedure_name" => "DROP_BINDING_CREDENTIAL",
            "drop_all_binding_credential_procedure_name" => "DROP_ALL_BINDING_CREDENTIALS",
            "validate_binding_credential_procedure_name" => "VALIDATE_BINDING_CREDENTIALS"
          }
        }
      }

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

    context "uses tls" do
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
    end
  end
end
