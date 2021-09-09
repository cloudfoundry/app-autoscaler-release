# Testing Autoscaler Bosh Release

To test Autoscaler Bosh release, we write Rspec test cases for a given job (component)
Here, component template files (.erb) are used to provide information to bosh. This information is required to create a autoscaler Bosh release.
Often, properties/values required by BOSH templating engine are sometimes hard to locate. Hence, resulting errors in test cases (Rspec).


Following are the most common problems/errors while writing Autoscaler release (created by BOSH)

## Identify required Properties of a particular job to be included in a test case

Assuming that a test case for scalingengine job (component):

- Go the template file(.erb) of the job e.g. scalingengine
- Find properties having string "scalingengine" and look for properties which do not have default key in it. Properties not having default should be included in the test case.
If these properties are not included, the test will not run. Therefore, inclusion of such properties/objects is required in the test case.

## - Undefined method `[]' for nil:NilClass

```
bundle exec rspec spec/jobs/scalingengine/scalingengine_spec.rb
```
**Output** 
```    
Failure/Error: rendered_template = YAML.safe_load(template.render(properties))
     
     NoMethodError:
       undefined method `[]' for nil:NilClass
     # (erb):32:in `get_binding'
     # /Users/<USER>/.rvm/gems/ruby-2.5.1/gems/bosh-template-2.2.1/lib/bosh/template/test/template.rb:23:in `render'
     # ./spec/jobs/scalingengine/scalingengine_spec.rb:108:in `block (3 levels) in <top (required)>'
```

#### Cause
 The find method called in template file (jobs/scalingengine/templates/scalingengine.yml.erb)  on line 32 (in the example above) is unable to find a certain property/object. Hence the error.


#### Resolution
- Go to the template file (erb) file for a particular job e.g. jobs/scalingengine/templates/scalingengine.yml.erb
- navigate to line number 32 and look for missing properties which needs to be included in the test case (spec/jobs/scalingengine/scalingengine_spec.rb)
In the example above, roles array is missing.

