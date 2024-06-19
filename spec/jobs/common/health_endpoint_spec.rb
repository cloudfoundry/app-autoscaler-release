require "rspec"
require "json"
require "bosh/template/test"
require "rspec/file_fixtures"
require "yaml"

describe "health endpoint sections relevant specs" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  [
    %w[apiserver golangapiserver config/apiserver.yml apiserver.yml],
    %w[eventgenerator eventgenerator config/eventgenerator.yml eventgenerator.yml],
    %w[metricsforwarder metricsforwarder config/metricsforwarder.yml metricsforwarder.yml],
    %w[operator operator config/operator.yml operator.yml],
    %w[scalingengine scalingengine config/scalingengine.yml scalingengine.yml]
  ].each do |service, release_job, config_file, properties_file|
    context service do
      context "health endpoint" do
        before(:each) do
          @properties = YAML.safe_load(fixture(properties_file).read)
          @template = release.job(release_job).template(config_file)
          @links = case service
          when "eventgenerator"
            [Bosh::Template::Test::Link.new(name: "eventgenerator")]
          else
            []
          end
          @rendered_template = YAML.safe_load(@template.render(@properties, consumes: @links))
        end
        it "by default TLS is not configured" do
          expect(@rendered_template["health"]["tls"]).to be_nil
        end

        it "TLS can be enabled" do
          service_config = (@properties["autoscaler"][service] ||= {})
          service_config["health"] = {
            "ca_cert" => "SOME_CA",
            "server_cert" => "SOME_CERT",
            "server_key" => "SOME_KEY"
          }

          rendered_template = YAML.safe_load(@template.render(@properties, consumes: @links))

          expect(rendered_template["health"]["tls"]).not_to be_nil
          expect(rendered_template["health"]["tls"]).to include({
            "key_file" => "/var/vcap/jobs/#{release_job}/config/certs/healthendpoint/server.key",
            "ca_file" => "/var/vcap/jobs/#{release_job}/config/certs/healthendpoint/ca.crt",
            "cert_file" => "/var/vcap/jobs/#{release_job}/config/certs/healthendpoint/server.crt"
          })
        end
      end
    end
  end
end
