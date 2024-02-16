# score-compose

**This is a POC to test out a potential future version of `score-compose`.**

## Usage

There are two commands available, `init` and `generate`. Init is used to set up
a new score compose context and must be called before the `generate` command is
used.

### Examples


This can be run in the root of this repository to stand up a single
score workload which depends on a postgres database and a redis.

```
score-compose init my-new-score-compose-project

score-compose generate

docker-compose up

```

## What is a Score Compose project?

`score-compose` allows for multiple score workloads to be added to a single
compose project (aka `compose.yaml` file.) This is achieved by allowing
the `generate` command to be used multiple times on different score files.

`score-compose` keeps track of the the latest invocation per score file of
the `generate` in the context. The context is stored in directory called
`.score-compose`. The context includes state such as generated resource
parameters so that they remain stable between invocations.


## Resource provisioners

`score-compose` has a template based approach to provisioning resources in
a compose project. This will be expanded in future to include running commands
or using web APIs.

### Template provisioning

Templates are defined in the
[provisioners.yaml](./internal/resources/provisioners.yaml) file. This file is
organised by resource type and defines a set of templates that evaluate to YAML
files.

The templating language used is the Go Templating language extended with
the [sprig](https://masterminds.github.io/sprig/) templating functions. These
are broadly the same functions available as in Helm Charts.

Each template is evaluated in order and is made available as inputs to other
templates, persisted, used as part of the compose config or causes files and/or
directories to be written to the file system.

## Template Inputs
Each template has the following set of inputs:
| Property | Type     | Description                                           |
| -------- | -------- | ----------------------------------------------------- |
| `id`     | `string` | a unique ID for the resource that is global to the    |
|          |          | context. It is guaranteed to be a RFC 1123 Label      |
|          |          | Name; https://tools.ietf.org/html/rfc1123             |
| `class`  | `string` | the class of the resource                             |
| `type`   | `string` | the type of the resource                              |
| `paths`  | `map`    | a map of paths useful for working with volumes.<br />`files`   - the directory into which the outputted files are written.<br /> `volumes` - the directory where volumes directories are created. |
| `params` | `map`    | any resource input parameters from the score file     |
| `state`  | `map`    | the last state that was stored for this resource      |
| `shared` | `map`    | the current shared global state                       |

Additional inputs depend on the template as enumerated in this table.
| tpl \ inputs | .init | .outputs | .services | .files | .volumeDirs |
| ------------ | ----- | -------- | --------- | ------ | ----------- |
| `init`       |       |          |           |        |             |
| `outputs`    |   x   |          |           |        |             |
| `files`      |   x   |     x    |           |        |             |
| `networks`   |   x   |     x    |           |        |             |
| `service`    |   x   |     x    |           |        |             |
| `volumeDirs` |   x   |     x    |           |        |             |
| `shared`     |   x   |     x    |     x     |   x    |      x      |
| `state`      |   x   |     x    |     x     |   x    |      x      |

### Templates

- `init`

  The output of this template is available to all other templates for this
  resource.

  Must evaluate to an object.

- `outputs`

  The output of this template are resource outputs that can be referenced using
  the `${resources.RESOURCENAME.OUTPUT}` style placeholders in the score file.

  Must evaluate to an object.

- `files`

  Each key in the output represents a relative file path. For each file path,
  the content in the value is written to the file system. The location is
  determined by prefixing `paths.files` to the relative path in the key.

  Must evaluate to an object where the values are strings. It is an error for
  the key to evaluate to an absolute path.

- `networks`

  Evaluates to a subset of the `networks` section in the compose config.

  Must evaluate to object where keys are network names and values are 
  [network](https://github.com/compose-spec/compose-spec/blob/master/06-networks.md)
  objects.

- `services`

  Evaluates to a subset of the `services` section in the compose config.

  Must evaluate to object where keys are network names and values are 
  [service](https://github.com/compose-spec/compose-spec/blob/master/05-services.md)
  objects.

- `volumeDirs`

  Each key in the output represents a relative directory path. For each
  directory path, a directory will be created if it does not already exist. The
  location is determined by prefixing `paths.volumes` to the relative path in
  the key.

  Must evaluate to an object. It is an error for the key to evaluate to an
  absolute path.

- `state`

  Used as the `state` input for all future invocations of templates for this
  resource.

  Must evaluate to an object.

- `shared`

  The output of this is merged with the current value of this using the JSON
  Merge Patch (https://datatracker.ietf.org/doc/html/rfc7386) strategy.
  Specifically, for each property in an object, the following happens
  recursively:

  - if a new property is present, it is added

  - if an existing property exists in both the old and new objects and the value of the property in the new object is a scalar or array, it is replaced

  - if an existing property 

  It is important to note that resulting merged object will be passed to
  templates of other resources possibly of the same type. The order that
  resources are provisioned is not defined, so care should be taken to make
  the state that is tracked in this object invarient to provisioning order.

  Must evaluate to an object.
