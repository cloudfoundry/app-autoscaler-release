require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "operator" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("operator") }
  let(:template) { job.template("config/operator.yml") }
  let(:properties) { YAML.safe_load(fixture("operator.yml").read) }

  context "config/operator.yml" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["operator"] = {
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

    it "check operator basic auth username and password" do
      properties["autoscaler"]["operator"] = {
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
  end
end
