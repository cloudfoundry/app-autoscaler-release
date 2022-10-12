def build_db_url(db_name, job_name)
  db_scheme = p("autoscaler.#{db_name}.db_scheme")
  db_address = p("autoscaler.#{db_name}.address")
  db_port = p("autoscaler.#{db_name}.port")
  db_role = p_arr("autoscaler.#{db_name}.roles").find { |role| role["tag"] == db_name.tr("_", "") or role["tag"] == "default" }
  db_database = p_arr("autoscaler.#{db_name}.databases").find { |database| database["tag"] == db_name.tr("_", "") or database["tag"] == "default" }

  if db_scheme == "postgres"
    db_url = "jdbc:postgresql://#{db_address}:#{db_port}/#{db_database["name"]}?ApplicationName=#{job_name}&sslmode=" + p("autoscaler.#{db_name}.sslmode")
  else
    db_url = "#{db_role['name']}:#{db_role['password']}@tcp(#{db_address}:#{db_port})/#{db_database['name']}?tls=" + p("autoscaler.#{db_name}.sslmode")
  end
  append_db_tls_configs(db_name, db_url, job_name)
end

def append_db_tls_configs(db_name, db_url, job_name)
  unless p("autoscaler.#{db_name}.tls.ca") == ""
    db_url = "#{db_url}&sslrootcert=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/ca.crt"
  end
  unless p("autoscaler.#{db_name}.tls.crt") == ""
    db_url = "#{db_url}&sslcert=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/crt"
  end
  unless p("autoscaler.#{db_name}.tls.key") == ""
    db_url = "#{db_url}&sslkey=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/key"
  end
  db_url
end
