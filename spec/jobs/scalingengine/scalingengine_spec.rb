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
        expect(rendered_template["health"]).to include({"username" => nil, "password" => nil})
      end

      it "does not include health port anymore" do
        expect(rendered_template["health"].keys).not_to include("port")
      end

      it "check scalingengine basic auth username and password" do
        properties["autoscaler"]["scalingengine"] = {
          "health" => {
            "username" => "test-user",
            "password" => "test-user-password"
          }
        }

        expect(rendered_template["health"]).to include(
          {
            "password" => "test-user-password"
          }
        )
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
