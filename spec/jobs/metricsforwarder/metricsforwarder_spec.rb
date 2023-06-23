require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "metricsforwarder" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("metricsforwarder") }
  let(:template) { job.template("config/metricsforwarder.yml") }
  let(:properties) { YAML.safe_load(fixture("metricsforwarder.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/metricsforwarder.yml" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["metricsforwarder"] = {
        "health" => {
          "port" => 1234
        }
      }

      expect(rendered_template["health"])
        .to include(
          {"port" => 1234}
        )
    end

    it "check metricsforwarder basic auth username and password" do
      properties["autoscaler"]["metricsforwarder"] = {
        "health" => {
          "port" => 1234,
          "username" => "test-user",
          "password" => "test-user-password"
        }
      }

      expect(rendered_template["health"])
        .to include(
          {"port" => 1234,
           "username" => "test-user",
           "password" => "test-user-password"}
        )
    end

    it "has a cred helper impl by default" do
      expect(rendered_template).to include(
        {
          "cred_helper_impl" => "default"
        }
      )
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

    context "health_server" do
      it "default port exist" do
        expect(rendered_template["health"]["port"]).to eq(6403)
      end

      it "credentials are defined" do
        expect(rendered_template["health"]["username"]).to eq("basic_auth_username")
        expect(rendered_template["health"]["password"]).to eq("basic_auth_secret")
      end

      it "readiness enabled is set to false by default" do
        expect(rendered_template["health"]["readiness_enabled"]).to eq(false)
      end

      it "unprotected_endpoint config is empty by default" do
        expect(rendered_template["health"]["unprotected_endpoints"]).to match_array([])
      end

      it "has valid endpoints in unprotected_endpoint config" do
        properties["autoscaler"]["metricsforwarder"]["health"]["unprotected_endpoints"] = %w[/debug/pprof /health/liveness /health/prometheus /health/readiness]
        expect(rendered_template["health"]["unprotected_endpoints"]).to contain_exactly("/health/liveness",
          "/health/prometheus",
          "/health/readiness",
          "/debug/pprof")
      end
    end
  end
end
