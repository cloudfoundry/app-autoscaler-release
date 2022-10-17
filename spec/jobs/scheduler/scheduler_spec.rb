require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "scheduler" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("scheduler") }
  let(:template) { job.template("config/application.properties") }
  let(:properties) { YAML.safe_load(fixture("scheduler.yml").read) }
  let(:rendered_template) { template.render(properties) }

  context "config/application.properties" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234
        }
      }
      rendered_template = template.render(properties)

      expect(rendered_template).to include("scheduler.healthserver.port=1234")
      expect(rendered_template).to include("scheduler.healthserver.basicAuthEnabled=false")
      expect(rendered_template).to include("scheduler.healthserver.username=")
      expect(rendered_template).to include("scheduler.healthserver.password=")
    end

    it "check scheduler username and password" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234,
          "basicAuthEnabled" => "true",
          "username" => "test-user",
          "password" => "test-user-password"
        }
      }
      rendered_template = template.render(properties)

      expect(rendered_template).to include("scheduler.healthserver.port=1234")
      expect(rendered_template).to include("scheduler.healthserver.basicAuthEnabled=true")
      expect(rendered_template).to include("scheduler.healthserver.username=test-user")
      expect(rendered_template).to include("scheduler.healthserver.password=test-user-password")
    end
  end
end
