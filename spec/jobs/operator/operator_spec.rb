require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "operator" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("operator") }
  let(:template) { job.template("config/operator.yml") }
  let(:properties) { YAML.safe_load(fixture("operator.yml").read) }
  let(:rendered_template) { YAML.safe_load(template.render(properties)) }

  context "config/operator.yml" do
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

      context "instance_metrics_db " do
        it "includes the ca, cert and key in url when configured" do
          rendered_template["instance_metrics_db"]["db"]["url"].tap do |url|
            check_if_certs_in_url(url, "instancemetrics_db")
          end
        end

        it "does not include the ca, cert and key in url when not configured" do
          properties["autoscaler"]["instancemetrics_db"]["tls"] = nil
          rendered_template["instance_metrics_db"]["db"]["url"].tap do |url|
            check_if_certs_not_in_url(url, "instancemetrics_db ")
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

    context "health_server" do
      it "default port exist" do
        expect(rendered_template["health"]["port"]).to eq(6208)
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
        properties["autoscaler"]["operator"]["health"]["unprotected_endpoints"] = %w[/debug/pprof /health/liveness /health/prometheus /health/readiness]
        expect(rendered_template["health"]["unprotected_endpoints"]).to contain_exactly("/health/liveness",
          "/health/prometheus",
          "/health/readiness",
          "/debug/pprof")
      end
    end
  end
end
