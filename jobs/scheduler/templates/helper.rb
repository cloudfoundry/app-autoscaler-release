def generate_db_url(db_name, job_name)
  db_scheme = p("autoscaler.#{db_name}.db_scheme")
  db_address = p("autoscaler.#{db_name}.address")
  db_port = p("autoscaler.#{db_name}.port")
  db_role = p_arr("autoscaler.#{db_name}.roles").find { |role| role['tag'] == db_name.tr('_', '') or role['tag'] == 'default' }
  db_database = p_arr("autoscaler.#{db_name}.databases").find { |database| database['tag'] == db_name.tr('_', '') or database['tag'] == 'default' }

  if db_scheme == "postgres"
    db_url = "jdbc:postgresql://#{db_address}:#{db_port}/#{db_database["name"]}?ApplicationName=#{job_name}&sslmode=" + p("autoscaler.#{db_name}.sslmode")

    unless p("autoscaler.#{db_name}.tls.ca") == ""
      db_url = "#{db_url}&sslrootcert=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/ca.crt"
    end
    unless p("autoscaler.#{db_name}.tls.crt") == ""
      db_url = "#{db_url}&sslcert=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/crt"
    end
    unless p("autoscaler.#{db_name}.tls.key") == ""
      db_url = "#{db_url}&sslkey=/var/vcap/jobs/#{job_name}/config/certs/#{db_name}/key"
    end

  else #mysql case
    db_url = "jdbc:mysql://" + db_address + ":" + db_port.to_s + "/" + db_database['name'] + "?autoReconnect=true"
    unless p('autoscaler.scheduler_db.tls.ca') == ""
      db_url = "#{db_url}&useSSL=true&requireSSL=true&verifyServerCertificate=true&enabledTLSProtocols=TLSv1.2&trustCertificateKeyStorePassword=123456&trustCertificateKeyStoreUrl=file:/var/vcap/data/certs/scheduler_db/cacerts&trustCertificateKeyStoreType=pkcs12"
    end
  end
  db_url
end
