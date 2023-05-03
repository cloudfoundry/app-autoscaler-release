require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "scheduler" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("scheduler") }
  let(:template) { job.template("config/scheduler.yml") }
  let(:properties) { YAML.safe_load(fixture("scheduler.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/scheduler.yml" do
    it "does set neither username nor password if not configured" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234,
          "unprotected_endpoints" => []
        }
      }

      rendered_template = YAML.safe_load(template.render(properties)) # klappt nicht

      expect(rendered_template).to include(
        {"scheduler" => {
          "healthserver" => {
            "port" => 1234,
            "username" => nil,
            "password" => nil,
            "unprotected_endpoints" => []
          }
        }}
      )
    end

    it "check scheduler username and password and allow access with basic auth" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234,
          "username" => "test-user",
          "password" => "test-user-password",
          "unprotected_endpoints" => ["/health/liveness"]
        }
      }

      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template).to include(
        {"scheduler" => {
          "healthserver" => {
            "port" => 1234,
            "username" => "test-user",
            "password" => "test-user-password",
            "unprotected_endpoints" => ["/health/liveness"]
          }
        }}
      )
    end

    it "extension properties are added to the properties file" do
      properties["autoscaler"]["scheduler"] = {
        "application" => {
          "props" => <<~HEREDOC
            logging:
              level:
                scheduler: "info"
                quartz: "info"
          HEREDOC
        }
      }

      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template).to include(
        {"logging" => {
          "level" => {
            "quartz" => "info",
            "scheduler" => "info"
          }
        }}
      )
    end
  end
end
