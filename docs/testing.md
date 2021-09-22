# Testing

When testing operator-builder, keep in mind you are testing a source code generator
that end-user engineers will use to manage Kubernetes operator projects.
Operator Builder is used to generate code for a distinct code repository - so the
testing is conducted as such.  It stamps out and/or modifies source code for an
operator project when a test is run.

At this time, manual verification of results is possible.  Additionally, the manual
verification steps have been place into a CI pipeline via the `.github/workflows`
directory.

There are several relevant make targets:

* `build`: Builds the operator-builder binary and saves it in the `bin`
  directory.
* `install`: Installs the operator-builder binary to `/usr/local/bin`.  **NOTE:**
  this will override any previous installations of operator-builder.
* `debug`: Runs the `delve` debugger in conjunction with the operator-builder codebase.
* `generate`: Builds operator-builder, and uses the built binary to run `init`
  and `create api` tasks into a `TEST_PATH` directory.
* `generate-clean`: Use with caution. Deletes the contents of the test repo directory.
* `debug-clean`: Use with caution. Deletes the contents of the debug repo directory.

Follow these steps to create a new test case:

1. Create a descriptive name for the test by creating a new directory under
   the `test/` directory.
2. Create a `.workloadConfig` directory within your newly created directory.
3. Add the YAML files for your workload under the newly created `.workloadConfig`
   directory.
4. Create a [workload configuration](workloads.md) in the `.workloadConfig` directory
   with the name `workload.yaml`.
5. In order to run the test in `test/application/.workloadConfig/`, run:
```bash
TEST_PATH=/tmp/test TEST_WORKLOAD_PATH=test/application make generate
```

   In order to run the test in `test/platform/.workloadConfig/`, run:
```bash
TEST_PATH=/path/to/generated/code TEST_WORKLOAD_PATH=test/platform make generate
```
3. To remove the generated files in the target test repo run:
```
make generate-clean
```
