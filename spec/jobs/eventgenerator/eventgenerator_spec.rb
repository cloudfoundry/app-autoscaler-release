require "rspec"
require "json"
require "bosh/template/test"
require "yaml"
require "rspec/file_fixtures"
require_relative "../utils"

describe "eventgenerator" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("eventgenerator") }
  let(:template) { job.template("config/eventgenerator.yml") }
  let(:properties) { YAML.safe_load(fixture("eventgenerator.yml").read) }

  context "config/eventgenerator.yml" do
    let(:links) do
      [
        Bosh::Template::Test::Link.new(
          name: "eventgenerator",
          properties: {}
        )
      ]
    end

    let(:rendered_template) { YAML.safe_load(template.render(properties, consumes: links)) }

    it "does not set username nor password if not configured" do
      properties["autoscaler"]["eventgenerator"] = {
        "health" => {
          "port" => 1234
        }
      }
      expect(rendered_template["health"]["server_config"]["port"]).to eq(properties["autoscaler"]["eventgenerator"]["health"]["port"])
    end

    it "check eventgenerator username and password" do
      properties["autoscaler"]["eventgenerator"] = {
        "health" => {
          "port" => 1234,
          "username" => "test-user",
          "password" => "test-user-password"
        }
      }

      expect(rendered_template["health"]["server_config"]["port"]).to eq(properties["autoscaler"]["eventgenerator"]["health"]["port"])
      expect(rendered_template["health"]["basic_auth"]["username"]).to eq(properties["autoscaler"]["eventgenerator"]["health"]["username"])
      expect(rendered_template["health"]["basic_auth"]["password"]).to eq(properties["autoscaler"]["eventgenerator"]["health"]["password"])
    end

    describe "when using log-cache via https/uaa" do
      before do
        properties["autoscaler"]["eventgenerator"] = {
          "metricscollector" => {
            "host" => "logcache.cf.test.com",
            "port" => "",
            "uaa" => {
              "client_id" => "logs_admin_client_id",
              "client_secret" => "logs_admin_client_secret",
              "url" => "uaa.cf.test.com"
            }
          }
        }
      end

      it "should not add metric_collector_port to metric_collector_url" do
        expect(rendered_template["metricCollector"]["metric_collector_url"])
          .not_to include(":")
      end

      it "should add uaa credentials" do
        expect(rendered_template["metricCollector"]["uaa"]).to include({
          "client_id" => "logs_admin_client_id",
          "client_secret" => "logs_admin_client_secret",
          "url" => "uaa.cf.test.com"
        })
      end

      it "should set uaa skip ssl validation to false by default" do
        expect(rendered_template["metricCollector"]["uaa"]["skip_ssl_validation"]).to be_falsey
      end
    end

    context "cf server" do
      it "includes default port for cf server" do
        expect(rendered_template["cf_server"]["port"]).to eq(8080)
      end

      it "defaults xfcc valid org and space " do
        properties["autoscaler"]["eventgenerator"] = {
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

      context "appmetrics_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["appmetrics_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "appmetrics_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["appmetrics_db"]["tls"] = nil
          rendered_template["db"]["appmetrics_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "appmetrics_db")
          end
        end
      end
    end
  end
end
