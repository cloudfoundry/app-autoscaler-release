require "rspec"
require "json"
require "bosh/template/test"
require "yaml"
require "rspec/file_fixtures"

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
      expect(rendered_template["health"])
        .to include(
          {"port" => 1234}
        )
    end

    it "check eventgenerator username and password" do
      properties["autoscaler"]["eventgenerator"] = {
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

    describe "use_log_cache feature" do
      it "keeps log cache off by default" do
        expect(rendered_template["metricCollector"])
          .to include({"use_log_cache" => false})
      end

      it "should add https protocol to metric_collector_url" do
        expect(rendered_template["metricCollector"]["metric_collector_url"])
          .to include("http")
      end

      describe "when log cache on" do
        before do
          properties["autoscaler"]["eventgenerator"] = {
            "metricscollector" => {
              "use_log_cache" => true
            }
          }
        end

        it "check eventgenerator use log cache" do
          expect(rendered_template["metricCollector"])
            .to include({"use_log_cache" => true})
        end

        it "should not add https protocol to metric_collector_url" do
          expect(rendered_template["metricCollector"]["metric_collector_url"])
            .not_to include("http")
        end
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

      context "appmetrics_db" do
        it "includes the ca, cert and key in url when configured" do
          puts rendered_template["db"]
          rendered_template["db"]["appmetrics_db"]["url"].tap do |url|
            expect(url).to include("sslrootcert=")
            expect(url).to include("appmetrics_db/ca.crt")
            expect(url).to include("sslkey=")
            expect(url).to include("appmetrics_db/key")
            expect(url).to include("sslcert=")
            expect(url).to include("appmetrics_db/crt")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["appmetrics_db"]["tls"] = nil
          rendered_template["db"]["appmetrics_db"]["url"].tap do |url|
            expect(url).to_not include("sslrootcert=")
            expect(url).to_not include("appmetrics_db/ca.crt")
            expect(url).to_not include("sslkey=")
            expect(url).to_not include("appmetrics_db/key")
            expect(url).to_not include("sslcert=")
            expect(url).to_not include("appmetrics_db/crt")
          end
        end
      end
    end
  end
end
