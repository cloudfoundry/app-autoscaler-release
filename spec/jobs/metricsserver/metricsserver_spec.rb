require "rspec"
require "json"
require "bosh/template/test"
require "yaml"
require "rspec/file_fixtures"

describe "metricsserver" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("metricsserver") }
  let(:template) { job.template("config/metricsserver.yml") }
  let(:properties) { YAML.safe_load(fixture("metricsserver.yml").read) }
  let(:links) { [Bosh::Template::Test::Link.new(name: "metricsserver")] }
  let(:rendered_template) { YAML.safe_load(template.render(properties, consumes: links)) }

  context "config/metricsserver.yml" do
    it "does not set username nor password if not configured" do
      properties["autoscaler"]["metricsserver"] = {
        "health" => {
          "port" => 1234
        }
      }

      expect(rendered_template["health"])
        .to include(
          {"port" => 1234}
        )
    end

    it "check metricsserver basic auth username and password" do
      properties["autoscaler"]["metricsserver"] = {
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
          rendered_template["db"]["policy_db"]["url"].tap do |url|
            expect(url).to include("sslrootcert=")
            expect(url).to include("policy_db/ca.crt")
            expect(url).to include("sslkey=")
            expect(url).to include("policy_db/key")
            expect(url).to include("sslcert=")
            expect(url).to include("policy_db/crt")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["policy_db"]["tls"] = nil
          rendered_template["db"]["policy_db"]["url"].tap do |url|
            expect(url).to_not include("sslrootcert=")
            expect(url).to_not include("policy_db/ca.crt")
            expect(url).to_not include("sslkey=")
            expect(url).to_not include("policy_db/key")
            expect(url).to_not include("sslcert=")
            expect(url).to_not include("policy_db/crt")
          end
        end
      end

      context "instancemetrics_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["instancemetrics_db"]["url"].tap do |url|
            expect(url).to include("sslrootcert=")
            expect(url).to include("instancemetrics_db/ca.crt")
            expect(url).to include("sslkey=")
            expect(url).to include("instancemetrics_db/key")
            expect(url).to include("sslcert=")
            expect(url).to include("instancemetrics_db/crt")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["instancemetrics_db"]["tls"] = nil
          rendered_template["db"]["instancemetrics_db"]["url"].tap do |url|
            expect(url).to_not include("sslrootcert=")
            expect(url).to_not include("instancemetrics_db/ca.crt")
            expect(url).to_not include("sslkey=")
            expect(url).to_not include("instancemetrics_db/key")
            expect(url).to_not include("sslcert=")
            expect(url).to_not include("instancemetrics_db/crt")
          end
        end
      end
    end
  end
end
