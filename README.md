### k8s-template
Provide a text golang template replacement based on an input map set.

I was running into use cases with multiple secrets or text replacement
options required for kubernetes files. This is intended to simplify
decoupling the secrets and dynamic configuration information from
public git commits and other exposures.


---
#### Create template definitions

Many secrets and shared configuration options need to be published
privately to cluster members.

Some, for example in the kubernetes secrets, need to be base64
encoded. Others are complete files, or environment variables which we
don't want to be exposed or committed to a git repo.

The definition of source files are de-coupled from the data for
secrets by defining templates and configuration values that are added
at configuration run time.

The execution of template options, is sequential. If a file that has
one or more templates to replace, is consumed/used in another step,
then the file must be produced/generated before the consumer/user of
that file can load it. 

Because this is an ordering dependency for file components which are
co-dependent and this approach, currently, doesn't account for those
options, or do circular dependency DAG analysis or similar, then those
co-dependent components must be manually sequenced.

Again, if some of these are components between private and public git
repos, there's friction between secrecy and data sharing, so the
generation of the components and promotion to cluster visibility is
dependent on the process, but assuming the access to network
components is limited to the cluster for insecure services and access
is limited to encrypted connections for edge ingress access, then
component sharing intra cluster is relatively secure.

---
#### name replacement description

*--mappings=template-replacements.yaml*

Describe source record replacement text information

A data struct exposed for text replacement using the golang template
process. You're responsible to ensure that the name field is uniq

An array of definitions for replacement items
Name: the map replacement name for the golang template format, 
   name: Repo 's value after interpolation would replace the template

    {{ .Repo }}

in a file


Use a standardized template of keyword key/value name/value pairs in a
rule file.

- Fields with base64: true have their values replaced by their base64
equivalent string

- fields with a file: true attribute are sourced from the named file

- fields with an env: true attribute are sourced from the named environment variable

If for some reason you need both a base64 version and a plain text
version of an attribute, you can apply the base64:false flag and use
the provided helper function to remap the value
 YamlConfig : base64: false

Then the plain text version can map via a direct:
    {{ .YamlConfig }}

and the helper function base64Encode can be used where needed
    {{ .YamlConfig | base64Encode }}

---
#### Example

0. Create a mapping.yaml file with an array of maps which define the values for template replacement
   - name: MappingName : the mapping name to use
   - value: replacement text or file or env variable 
   - base64: base64 encode the final value 
   - file: source the value from the using 'value' as the file name
   - env: source the value from the environment using 'value' as the env var name
0. Modify the secrets or other text replacement fields in the definition with {{ .MappingName }}
0. Optionally disable the base64 flag, but use {{ .MappingName | base64Encode }} 

*YAML Map describes the mapping*

```
- name: Text
  base64: true
  value: Replacement Text

- name: PublicKey
  base64: false
  file: true
  value: ~/.ssh/id_rsa.pub
```

*Template description of the golang mapping*

```
apiVersion: v1
kind: Secret
metadata:
  namespace: default
  name: cluster-config-secret
data:
# already base64 encoded
  text: {{ .Text }}
# base64 encode for the secret 
  id-rsa.pub: {{ .PublicKey | base64Encode }}
```

---
#### Example execution

*bin/k8s-template*

- ```bin/k8s-template --template=template.yaml --mappings=mappings.yaml```


---
#### Known issue [ based on os.Getenv and some environments ]


