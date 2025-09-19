# Read-me #

The currently used golang-library for json-schema, namely “[gojsonschema](<https://github.com/xeipuuv/gojsonschema>)” not seems to be capable to resolve references to other files correctly across different directories. Perhaps not limited to but especially when referencing up the file-system-hierarchy (parent-directories). *Hint*: Use symbolic links to circumvent the issue!
