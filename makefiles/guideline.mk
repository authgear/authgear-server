# Guideline on authoring makefiles.
#
# 1. We use simply expanded variables.
# This means
# - You use ::= instead of = because = defines a recursively expanded variable.
#   See https://www.gnu.org/software/make/manual/html_node/Simple-Assignment.html
# - You use ::= instead of := because ::= is a POSIX standard.
#   See https://www.gnu.org/software/make/manual/html_node/Simple-Assignment.html
# - You do not use ?= because it is shorthand to define a recursively expanded variable.
#   See https://www.gnu.org/software/make/manual/html_node/Conditional-Assignment.html
#   You should use the long form documented in the above link instead.
# - When you override a variable in the command line, as documented in https://www.gnu.org/software/make/manual/html_node/Overriding.html
#   you specify the variable with ::= instead of = or :=
#   If you fail to do so, the variable becomes recursively expanded variable accidentally.
#
# 2. Global variables SHOULD BE declared at the beginning of the makefile.
#    Global variables are available to all makefile targets.
#
# 3. Target-specific variables SHOULD BE used if the variables are applicable to 1 target only.
#    See https://www.gnu.org/software/make/manual/html_node/Target_002dspecific.html
#
# 4. Makefile conditionals controls what makefile "sees".
#    Therefore, the body usually does not have leading tab characters.
#    See https://www.gnu.org/software/make/manual/html_node/Conditionals.html
#
# 5. Environment variables seen by make become a a make variable automatically.
#    See https://www.gnu.org/software/make/manual/html_node/Environment.html
#    Therefore, you DO NOT write `make something A="$A" B="$B"`.
#    Instead, you should make the environment variable visible to make.
#    For example,
#
#      export A=A
#      export B=B
#      make something
#
#    or
#
#      A=A B=B make something
