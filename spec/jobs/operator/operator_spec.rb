require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"
require_relative "../utils"

describe "operator" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("operator") }
  let(:template) { job.template("config/operator.yml") }
  let(:properties) { YAML.safe_load(fixture("operator.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/operator.yml" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["operator"] = {
        "health" => {
          "port" => 1234
        }
      }

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

      expect(rendered_template["health"])
        .to include(
          {"port" => 1234,
           "username" => "test-user",
           "password" => "test-user-password"}
        )
    end

    context "uses tls" do
      context "policy_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["app_syncer"]["db"]["url"].tap do |url|
            check_if_certs_in_url(url, "policy_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["policy_db"]["tls"] = nil
          rendered_template["app_syncer"]["db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "policy_db")
          end
        end
      end

      context "app_metrics_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["app_metrics_db"]["db"]["url"].tap do |url|
            check_if_certs_in_url(url, "appmetrics_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["appmetrics_db"]["tls"] = nil
          rendered_template["app_metrics_db"]["db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "appmetrics_db")
          end
        end
      end

      context "scaling_engine_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["scaling_engine_db"]["db"]["url"].tap do |url|
            check_if_certs_in_url(url, "scalingengine_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scalingengine_db"]["tls"] = nil
          rendered_template["scaling_engine_db"]["db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "scalingengine_db")
          end
        end
      end

      context "db_lock" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db_lock"]["db"]["url"].tap do |url|
            check_if_certs_in_url(url, "lock_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["lock_db"]["tls"] = nil
          rendered_template["db_lock"]["db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "lock_db")
          end
        end
      end
    end
  end
end
