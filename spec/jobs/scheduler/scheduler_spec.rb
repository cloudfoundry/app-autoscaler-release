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

      expect(rendered_template).to include("scheduler.healthserver.port=1234")
      expect(rendered_template).to include("scheduler.healthserver.basicAuthEnabled=true")
      expect(rendered_template).to include("scheduler.healthserver.username=test-user")
      expect(rendered_template).to include("scheduler.healthserver.password=test-user-password")
    end

    context "uses tls" do
      context "policy_db" do
        it "includes the ca, cert and key in url when configured" do
          expect(rendered_template).to include("sslrootcert=")
          expect(rendered_template).to include("policy_db/ca.crt")
          expect(rendered_template).to include("sslkey=")
          expect(rendered_template).to include("policy_db/key")
          expect(rendered_template).to include("sslcert=")
          expect(rendered_template).to include("policy_db/crt")
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["policy_db"]["tls"] = nil
          expect(rendered_template).to_not include("policy_db/ca.crt")
          expect(rendered_template).to_not include("policy_db/key")
          expect(rendered_template).to_not include("policy_db/crt")
        end
      end

      context "scheduler_db" do
        it "includes the ca, cert and key in url when configured" do
          expect(rendered_template).to include("sslrootcert=")
          expect(rendered_template).to include("scheduler_db/ca.crt")
          expect(rendered_template).to include("sslkey=")
          expect(rendered_template).to include("scheduler_db/key")
          expect(rendered_template).to include("sslcert=")
          expect(rendered_template).to include("scheduler_db/crt")
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["scheduler_db"]["tls"] = nil
          expect(rendered_template).to_not include("scheduler_db/ca.crt")
          expect(rendered_template).to_not include("scheduler_db/key")
          expect(rendered_template).to_not include("scheduler_db/crt")
        end
      end
    end
  end
end
