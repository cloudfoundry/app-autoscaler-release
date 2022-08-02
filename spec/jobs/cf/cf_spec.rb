require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "cf sections relevant specs" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  [
    %w[api_server golangapiserver config/apiserver.yml apiserver.yml],
    %w[scalingengine scalingengine config/scalingengine.yml scalingengine.yml],
    %w[operator operator config/operator.yml operator.yml]
  ].each do |service, release_job, config_file, properties_file|
    context service do
      context "cf" do
        before(:each) do
          @properties = YAML.safe_load(fixture(properties_file).read)
          @template = release.job(release_job).template(config_file)
          @rendered_template = YAML.safe_load(@template.render(@properties))
        end
        it "should have all the settings" do
          expect(@rendered_template["cf"]).to eq(
            {
              "api" => "https://api.#{service}.domain",
              "client_id" => "client_id",
              "secret" => "uaa_secret",
              # default
              "max_retries" => 3,
              # default
              "max_retry_wait_ms" => 0,
              # default
              "skip_ssl_validation" => false
            }
          )
        end
        context "check setting items that default" do
          it "max_retries" do
            @properties["autoscaler"]["cf"]["max_retries"] = 23
            rendered_template = YAML.safe_load(@template.render(@properties))
            expect(rendered_template["cf"]).to include({"max_retries" => 23})
          end
          it "skip_ssl_validation" do
            @properties["autoscaler"]["cf"]["skip_ssl_validation"] = false
            rendered_template = YAML.safe_load(@template.render(@properties))
            expect(rendered_template["cf"]).to include({"skip_ssl_validation" => false})
          end
          it "max_retry_wait_ms" do
            @properties["autoscaler"]["cf"]["max_retry_wait_ms"] = 3
            rendered_template = YAML.safe_load(@template.render(@properties))
            expect(rendered_template["cf"]).to include({"max_retry_wait_ms" => 3})
          end
        end
      end
    end
  end
end
