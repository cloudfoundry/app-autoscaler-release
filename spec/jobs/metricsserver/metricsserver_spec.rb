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

      context "instance_metrics_db" do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["db"]["instance_metrics_db"]["url"].tap do |url|
            check_if_certs_in_url(url, "instancemetrics_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["instancemetrics_db"]["tls"] = nil
          rendered_template["db"]["instance_metrics_db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "instancemetrics_db")
          end
        end
      end
    end

    context "health_server" do
      it "default port exist" do
        expect(rendered_template["health"]["port"]).to eq(6303)
      end

      it "credentials are defined" do
        expect(rendered_template["health"]["username"]).to eq("basic_auth_username")
        expect(rendered_template["health"]["password"]).to eq("basic_auth_secret")
      end

      it "readiness enabled is set to false by default" do
        expect(rendered_template["health"]["readiness_enabled"]).to eq(false)
      end

      it "unprotected_endpoint config is empty by default" do
        expect(rendered_template["health"]["unprotected_endpoints"]).to match_array([])
      end

      it "has valid endpoints in unprotected_endpoint config" do
        properties["autoscaler"]["metricsserver"]["health"]["unprotected_endpoints"] = %w[/debug/pprof /health/liveness /health/prometheus /health/readiness]
        expect(rendered_template["health"]["unprotected_endpoints"]).to contain_exactly("/health/liveness",
          "/health/prometheus",
          "/health/readiness",
          "/debug/pprof")
      end
    end
  end
end
