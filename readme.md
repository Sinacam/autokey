# Autokey
Autokey is a simple tool that allows you to manipulate mouse and keyboard inputs according to a human friendly configuration file.

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