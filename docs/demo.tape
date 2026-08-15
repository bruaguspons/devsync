# VHS tape for the devsync demo GIF — regenerated on every release by
# .github/workflows/release.yml (job: demo-gif). Do not hand-edit
# docs/demo.gif; edit this script instead.
#
# The demo opens with the real install one-liner, run for real against the
# release that was just published for this tag ($HOME/.local/bin ends up on
# PATH with the genuine downloaded binary). Right after that, a Hide/Show
# block silently swaps the just-installed mise.toml.snapshot/skills.snapshot
# (this repo's real, noisy ones) for a small curated fixture, so the rest of
# the recording keeps showing a clean, readable diff:
#   - ~/.local/share/devsync/mise.toml.snapshot pinning a lightweight tool
#   - ~/.local/share/devsync/skills.snapshot with one demo skill
#   - an empty ~/.claude/skills (so the skill diffs as "install")
# $FIXTURE_DIR (holding that curated fixture) and $HOME are set by the
# demo-gif job before vhs runs. See that job for how both are built.
#
# Docs: https://github.com/charmbracelet/vhs

Output docs/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1000
Set Height 560
Set Theme "Dracula"
Set Padding 20

Set TypingSpeed 40ms

Type "curl -fsSL https://raw.githubusercontent.com/bruaguspons/devsync/main/install.sh | bash"
Enter
Sleep 5s

Hide
Type "cp -f $FIXTURE_DIR/mise.toml.snapshot $HOME/.local/share/devsync/mise.toml.snapshot && rm -rf $HOME/.local/share/devsync/skills.snapshot $HOME/.claude/skills && cp -r $FIXTURE_DIR/skills.snapshot $HOME/.local/share/devsync/skills.snapshot && mkdir -p $HOME/.claude/skills && clear"
Enter
Sleep 500ms
Show

Set TypingSpeed 60ms

Type "devsync"
Enter
Sleep 1500ms

Space
Sleep 700ms
Down
Sleep 400ms
Space
Sleep 700ms
Enter
Sleep 1200ms

Left
Sleep 700ms
Enter
Sleep 3s

Type "devsync"
Enter
Sleep 1500ms
