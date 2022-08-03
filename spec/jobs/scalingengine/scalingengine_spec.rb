require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "scalingengine" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("scalingengine") }
  let(:template) { job.template("config/scalingengine.yml") }
  let(:properties) { YAML.safe_load(fixture("scalingengine.yml").read) }

  context "config/scalingengine.yml" do
    context "scalingengine" do
      it "does not set username nor password if not configured" do
        properties["autoscaler"]["scalingengine"] = {
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

      it "check scalingengine basic auth username and password" do
        properties["autoscaler"]["scalingengine"] = {
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
end
