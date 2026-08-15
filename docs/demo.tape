# VHS tape for the devsync demo GIF — regenerated on every release by
# .github/workflows/release.yml (job: demo-gif). Do not hand-edit
# docs/demo.gif; edit this script instead.
#
# Assumes the environment already has, isolated under $HOME:
#   - a devsync binary on PATH
#   - ~/.local/share/devsync/mise.toml.snapshot pinning a lightweight tool
#   - ~/.local/share/devsync/skills.snapshot with one demo skill
#   - an empty ~/.claude/skills (so the skill diffs as "install")
# See the demo-gif job for how that fixture is built.
#
# Docs: https://github.com/charmbracelet/vhs

Output docs/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1000
Set Height 560
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 60ms

Type "devsync"
Enter
Sleep 1500ms

Space
Down
Space
Enter
Sleep 500ms

Left
Enter
Sleep 6s

Type "devsync"
Enter
Sleep 1500ms
