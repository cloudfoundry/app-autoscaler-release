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

  context "config/metricsforwarder.yml" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["metricsforwarder"] = {
        "health" => {
          "port" => 1234
        }
      }
      rendered_template = YAML.safe_load(template.render(properties))

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
      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template["health"])
        .to include(
          {"port" => 1234,
           "username" => "test-user",
           "password" => "test-user-password"}
        )
    end

    it "has a cred helper impl by default" do
      rendered_template = YAML.safe_load(template.render(properties))
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
end
