def check_if_certs_in_url(url, db_name)
  expect(url).to include("sslrootcert=")
  expect(url).to include("#{db_name}/ca.crt")
  expect(url).to include("sslkey=")
  expect(url).to include("#{db_name}/key")
  expect(url).to include("sslcert=")
  expect(url).to include("#{db_name}/crt")
end

def check_if_certs_not_in_url(url, db_name)
  expect(url).to_not include("sslrootcert=")
  expect(url).to_not include("#{db_name}/ca.crt")
  expect(url).to_not include("sslkey=")
  expect(url).to_not include("#{db_name}/key")
  expect(url).to_not include("sslcert=")
  expect(url).to_not include("#{db_name}/crt")
end
