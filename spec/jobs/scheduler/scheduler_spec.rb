require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"
require_relative "../utils"

describe "scheduler" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("scheduler") }
  let(:template) { job.template("config/scheduler.yml") }
  let(:properties) { YAML.safe_load(fixture("scheduler.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "cf server" do
    it "default http port set to 0 - disabled" do
      expect(rendered_template["server"]["http"]["port"]).to eq(0)
    end

    it "defaults xfcc valid org and space" do
      properties["autoscaler"]["scheduler"] = {}
      properties["autoscaler"]["scheduler"]["cf_server"] = {}
      properties["autoscaler"]["scheduler"]["cf_server"]["xfcc"] = {
        "valid_org_guid" => "some-valid-org-guid",
        "valid_space_guid" => "some-valid-space-guid"
      }

      expect(rendered_template["cfserver"]["validOrgGuid"]).to eq(properties["autoscaler"]["scheduler"]["cf_server"]["xfcc"]["valid_org_guid"])
      expect(rendered_template["cfserver"]["validSpaceGuid"]).to eq(properties["autoscaler"]["scheduler"]["cf_server"]["xfcc"]["valid_space_guid"])
    end
  end

  context "Health Configuration" do
    it "does set neither username nor password if not configured" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234
        }
      }

      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template).to include(
        {"scheduler" => {
          "healthserver" => {
            "port" => 1234,
            "username" => "",
            "password" => "",
            "basicAuthEnabled" => false
          }
        }}
      )
    end

    it "check scheduler username and password and allow access with basic auth" do
      properties["autoscaler"]["scheduler"] = {
        "health" => {
          "port" => 1234,
          "username" => "test-user",
          "password" => "test-user-password"
        }
      }

      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template).to include(
        {"scheduler" => {
          "healthserver" => {
            "port" => 1234,
            "username" => "test-user",
            "password" => "test-user-password",
            "basicAuthEnabled" => false
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

  context "Datasource Configuration" do
    it "verify database username and password have string types" do
      rendered_template = YAML.safe_load(template.render(properties))

      expect(rendered_template["spring"]["datasource"]["username"]).to be_kind_of(String)
      expect(rendered_template["spring"]["datasource"]["username"]).not_to be_kind_of(Float)
      expect(rendered_template["spring"]["datasource"]["username"]).not_to eq(2222e123)
      expect(rendered_template["spring"]["datasource"]["username"]).to eq("2222e123")

      expect(rendered_template["spring"]["datasource"]["password"]).to be_kind_of(String)
      expect(rendered_template["spring"]["datasource"]["password"]).not_to be_kind_of(Float)
      expect(rendered_template["spring"]["datasource"]["password"]).to eq("default")
    end
  end
end
