# Testing

When testing operator-builder, keep in mind you are testing a source code generator
that end-user engineers will use to manage Kubernetes operator projects.
Operator Builder is used to generate code for a distinct code repository - so the
testing is conducted as such.  It stamps out and/or modifies source code for an
operator project when a test is run.

At this time, manual verification of results is required.  In the future,
functional tests for the resulting operator will be added.

There are 3 relevant make targets:

* `build`: Builds the operator-builder binary and saves it in the `bin`
  directory.
* `test-install`: Builds operator-builder and installs it at `/usr/local/bin/`.
* `test`: Builds and installs operator-builder, copies the secified test script
  to the `.test` directory to your test repo and runs that script.
* `test-clean`: Use with caution. Deletes the contents of the test repo directory.

Follow these steps to create a new test case:

1. Add a bash script to the `test/` directory that writes test config and
   manifest files and then performs the actions you would expect an end-user
   engineer to perform when using operator-builder.  See the existing scripts
   for examples.
2. In order to run the test in `test/application/.workloadConfig/`, run:
```bash
TEST_PATH=/my/test/repo/path TEST_SCRIPT=application.sh make test
```

   In order to run the test in `test/platform/.workloadConfig/`, run:
```bash
TEST_PATH=/my/test/repo/path TEST_SCRIPT=platform.sh make test
```
3. To remove the generated files in the target test repo run:
```
TEST_PATH=/my/test/repo/path TEST_SCRIPT=platform.sh make test-clean
```
