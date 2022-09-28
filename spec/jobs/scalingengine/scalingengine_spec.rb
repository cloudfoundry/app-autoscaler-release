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
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/scalingengine.yml" do
    context "scalingengine" do
      it "does not set username nor password if not configured" do
        properties["autoscaler"]["scalingengine"] = {
          "health" => {
            "port" => 1234
          }
        }

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

        expect(rendered_template["health"])
          .to include(
            {"port" => 1234,
             "username" => "test-user",
             "password" => "test-user-password"}
          )
      end
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

      context "scalingengine_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["scalingengine_db"]["url"].tap do |url|
            expect(url).to include("sslrootcert=")
            expect(url).to include("scalingengine_db/ca.crt")
            expect(url).to include("sslkey=")
            expect(url).to include("scalingengine_db/key")
            expect(url).to include("sslcert=")
            expect(url).to include("scalingengine_db/crt")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scalingengine_db"]["tls"] = nil
          rendered_template["db"]["scalingengine_db"]["url"].tap do |url|
            expect(url).to_not include("sslrootcert=")
            expect(url).to_not include("scalingengine_db/ca.crt")
            expect(url).to_not include("sslkey=")
            expect(url).to_not include("scalingengine_db/key")
            expect(url).to_not include("sslcert=")
            expect(url).to_not include("scalingengine_db/crt")
          end
        end
      end

      context "scheduler_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["scheduler_db"]["url"].tap do |url|
            expect(url).to include("sslrootcert=")
            expect(url).to include("scheduler_db/ca.crt")
            expect(url).to include("sslkey=")
            expect(url).to include("scheduler_db/key")
            expect(url).to include("sslcert=")
            expect(url).to include("scheduler_db/crt")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scheduler_db"]["tls"] = nil
          rendered_template["db"]["scheduler_db"]["url"].tap do |url|
            expect(url).to_not include("sslrootcert=")
            expect(url).to_not include("scheduler_db/ca.crt")
            expect(url).to_not include("sslkey=")
            expect(url).to_not include("scheduler_db/key")
            expect(url).to_not include("sslcert=")
            expect(url).to_not include("scheduler_db/crt")
          end
        end
      end
    end
  end
end
