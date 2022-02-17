# Autokey
Autokey is a simple tool that allows you to manipulate mouse and keyboard inputs according to a human friendly config file.

## Examples
### Hold Left Click
```yaml
hold: left click
```
In case you can't find something to hold down your mouse.

### Key Macros
```yaml
do:
  on: f6
  press: [right, down, right, a]
```
Performs ~~shoryuken~~ a series of key presses when you press `f6`.

### Key Spamming
```yaml
do:
  on: right click
  repeat:
    at: 5hz
    for: 1s
    press: d
```
Spams `d` key for one second at a frequency of 5hz when you right click.

### More Key Spamming
```yaml
do:
  on: f6
  repeat:
    at: 10hz
    until: f7
    press: left click
```
Spam left click when you press `f6` until you press `f7`.

## Config Syntax
The config file is a [yaml][1] file, which is composed of mappings, sequences and scalars. Only certain keys in mappings are allowed, specifying either an action or a description of the action. For example,
```yaml
press: a
```
`press` is an action, and its value is what key it will press.
```yaml
do:
  on: b
  press: a
```
`do` is an action that has a description `on` specifying a trigger. Other actions (i.e. `press`) are performed when it triggers.

Similarly, `repeat` is an action that has descriptions of how frequently, how long or when it stops triggering other actions.
```yaml
repeat:
  at: 2hz
  for: 3s
  press: a
```

Actions may be nested
```yaml
do:
  on: right click
  repeat:
    at: 5hz
    for: 1s
    press: d
```

In most cases, the values used are just a string (e.g. `5hz`). Sequences (e.g. `[a, b, c]`) are allowed for certain values where it makes sense. A top level sequence of actions can be used to execute multiple actions
```yaml
- do:
    on: a
    press: b
- do:
    on: c
    press: d
```

## Actions
### `do`
`do` is used to specify triggers by `on`. `on` may be a sequence, meaning it will be triggered by _any_ element. If no `on` is specified, `do` simply executes the nested actions.

### `repeat`
`repeat` repeatedly executes nested actions at a frequency specified by `at`. It ends either after a time specified by `for`, or until triggered by the key specified by `until`. `until` may be a sequence, meaning it will be triggered by _any_ element.

### `press`
Press and release the specified key. A key may be suffixed with `up` or `down`, meaning the key will be only be held down or released. `press` may be a sequence, meaning it will press the keys in order then release them in order.

### `hold`
Holds the specified key. Same as `press` with keys sufixed with `down`.

### `release`
Releases the specified key. Same as `press` with keys sufixed with `up`.

### `file`
Treats the content of the specified file as if it were in place of `file`.



[1]: https://www.cloudbees.com/blog/yaml-tutorial-everything-you-need-get-started