require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"
require_relative "../utils"

describe "scalingengine" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("scalingengine") }
  let(:template) { job.template("config/scalingengine.yml") }
  let(:properties) { YAML.safe_load(fixture("scalingengine.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/scalingengine.yml" do
    context "scalingengine" do
      it "does not set username nor password if not configured" do
        properties["autoscaler"]["scalingengine"] = {
          "health" => {
            "port" => 1234
          }
        }

        expect(rendered_template["health"]["server_config"]["port"]).to eq(properties["autoscaler"]["scalingengine"]["health"]["port"])
      end

      it "check scalingengine basic auth username and password" do
        properties["autoscaler"]["scalingengine"] = {
          "health" => {
            "port" => 1234,
            "username" => "test-user",
            "password" => "test-user-password"
          }
        }

        expect(rendered_template["health"]["server_config"]["port"]).to eq(properties["autoscaler"]["scalingengine"]["health"]["port"])
        expect(rendered_template["health"]["basic_auth"]["username"]).to eq(properties["autoscaler"]["scalingengine"]["health"]["username"])
        expect(rendered_template["health"]["basic_auth"]["password"]).to eq(properties["autoscaler"]["scalingengine"]["health"]["password"])
      end
    end

    context "cf server" do
      it "includes default port for cf server" do
        expect(rendered_template["cf_server"]["port"]).to eq(8080)
      end

      it "defaults xfcc valid org and space " do
        properties["autoscaler"]["scalingengine"] = {
          "cf_server" => {
            "xfcc" => {
              "valid_org_guid" => "some-valid-org-guid",
              "valid_space_guid" => "some-valid-space-guid"
            }
          }
        }

        expect(rendered_template["cf_server"]["xfcc"]).to include({
          "valid_org_guid" => "some-valid-org-guid",
          "valid_space_guid" => "some-valid-space-guid"
        })
      end
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

      context "scalingengine_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["scalingengine_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "scalingengine_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scalingengine_db"]["tls"] = nil
          rendered_template["db"]["scalingengine_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "scalingengine_db")
          end
        end
      end

      context "scheduler_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["scheduler_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "scheduler_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scheduler_db"]["tls"] = nil
          rendered_template["db"]["scheduler_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "scheduler_db")
          end
        end
      end
    end
  end
end
