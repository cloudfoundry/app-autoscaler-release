require "rspec"
require "json"
require "bosh/template/test"
require "yaml"

describe "eventgenerator" do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), "../../..")) }
  let(:job) { release.job("eventgenerator") }
  let(:template) { job.template("config/eventgenerator.yml") }
  let(:properties) do
    YAML.safe_load(%(
      autoscaler:
        policy_db:
          address: 10.11.137.101
          databases:
          - name: foo
            password: default
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
        appmetrics_db:
          address: 10.11.137.101
          databases:
          - name: foo
            password: default
            tag: default
          db_scheme: postgres
          port: 5432
          roles:
          - name: foo
            password: default
            tag: default
    ))
  end

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
        expect(rendered_template['metricCollector'])
          .to include({"use_log_cache" => false })
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
          expect(rendered_template['metricCollector'])
            .to include({"use_log_cache" => true })
        end
      end
    end
  end
end
