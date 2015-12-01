### k8s-template
Provide a text golang template replacement based on an input map set.

I was running into use cases with multiple secrets or text replacement
options required for kubernetes files. This is intended to simplify
decoupling the secrets and dynamic configuration information from
public git commits and other exposures.


---
#### mappings.yaml

Describe source record replacement text information

A data struct exposed for text replacement using the golang template
process. You're responsible to ensure that the name field is uniq

An array of definitions for replacement items
Name: the map replacement name for the golang template format, 
   name: Repo 's value after interpolation would replace the template
        {{ .Repo }}
in a file


Trivial first pass, use a standardized template of keyword key/value
name/value pairs in a rule file.

Fields with base64: true have their values replaced by their base64
equivalent string

fields with a file: true attribute are sourced from the named file

fields with an env: true attribute are sourced from the named environment variable

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
