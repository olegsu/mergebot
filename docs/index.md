# Pull Requests Robot

Chatops over github issues and pull requests.

## Config
The default prefix is `/bot`. You can control it by adding a `.prbot.yaml` file to the root of the repository`

### Example
```yaml
version: 1.0.0
use: bot # change me 
```

### Commands

* `/? help`
* `/? label {name}` - adds a label, creating new one if not exists
* `/? merge` - squash merge the pull request
* `/? workflow {name} {key=value} {key=value}` - uses workflow dispatch event api to trigger worklfow. The workflow must have `on: workflow_dispatch`.
  * `{name}` the file name of the worklfow (release.yaml)
  * `{key=value}` key and value pairs to pass as an input to the workflow

### Permissions
Only a commands from the repository owner will be accepted.
Permission feature can be found [here](https://github.com/olegsu/pull-requests-robot/issues/8) for feedback and suggestions.

