package yaml_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/yaml-workflow/yaml"
	"github.com/stretchr/testify/require"
)

func ExampleCreateStep_nestedObject() {
	pcore.Do(func(ctx px.Context) {
		ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
		workflowFile := "testdata/tf-k8s-sample.yaml"
		content, err := ioutil.ReadFile(workflowFile)
		if err != nil {
			panic(err.Error())
		}
		a := yaml.CreateStep(ctx, workflowFile, content)

		sb := service.NewServiceBuilder(ctx, `Yaml::Test`)
		sb.RegisterStateConverter(yaml.ResolveState)
		sb.RegisterStep(a)
		sv := sb.Server()
		_, defs := sv.Metadata(ctx)

		wf := defs[0]
		ac, _ := wf.Properties().Get4(`steps`)
		rs := ac.(px.List).At(0).(serviceapi.Definition)

		st := sv.State(ctx, rs.Identifier().Name(), px.EmptyMap)
		st.ToString(os.Stdout, px.Pretty, nil)
		fmt.Println()
	})

	// Output:
	// Kubernetes::Namespace(
	//   'metadata' => {
	//     'name' => 'terraform-lyra',
	//     'resource_version' => 'hi',
	//     'self_link' => 'me'
	//   },
	//   'namespace_id' => 'ignore'
	// )
}

func ExampleCreateStep() {
	pcore.Do(func(ctx px.Context) {
		ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
		workflowFile := "testdata/aws_vpc.yaml"
		content, err := ioutil.ReadFile(workflowFile)
		if err != nil {
			panic(err.Error())
		}
		a := yaml.CreateStep(ctx, workflowFile, content)

		sb := service.NewServiceBuilder(ctx, `Yaml::Test`)
		sb.RegisterStateConverter(yaml.ResolveState)
		sb.RegisterStep(a)
		sv := sb.Server()
		_, defs := sv.Metadata(ctx)

		wf := defs[0]
		wf.ToString(os.Stdout, px.Pretty, nil)
		fmt.Println()

		st := sv.State(ctx, `aws_vpc::vpc`, px.Wrap(ctx, map[string]interface{}{
			`tags`: map[string]string{`a`: `av`, `b`: `bv`}}).(px.OrderedMap))
		st.ToString(os.Stdout, px.Pretty, nil)
		fmt.Println()
	})

	// Output:
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'definition',
	//     'name' => 'aws_vpc'
	//   ),
	//   'serviceId' => TypedName(
	//     'namespace' => 'service',
	//     'name' => 'Yaml::Test'
	//   ),
	//   'properties' => {
	//     'parameters' => [
	//       Lyra::Parameter(
	//         'name' => 'tags',
	//         'type' => Hash[String, String],
	//         'value' => Deferred(
	//           'name' => 'lookup',
	//           'arguments' => ['aws.tags']
	//         )
	//       )],
	//     'returns' => [
	//       Lyra::Parameter(
	//         'name' => 'vpcId',
	//         'type' => String
	//       ),
	//       Lyra::Parameter(
	//         'name' => 'subnetId',
	//         'type' => String
	//       )],
	//     'steps' => [
	//       Service::Definition(
	//         'identifier' => TypedName(
	//           'namespace' => 'definition',
	//           'name' => 'aws_vpc::vpc'
	//         ),
	//         'serviceId' => TypedName(
	//           'namespace' => 'service',
	//           'name' => 'Yaml::Test'
	//         ),
	//         'properties' => {
	//           'parameters' => [
	//             Lyra::Parameter(
	//               'name' => 'tags',
	//               'type' => Hash[String, String]
	//             )],
	//           'returns' => [
	//             Lyra::Parameter(
	//               'name' => 'vpcId',
	//               'type' => Optional[String]
	//             )],
	//           'resourceType' => Aws::Vpc,
	//           'style' => 'resource',
	//           'origin' => '(file: testdata/aws_vpc.yaml)'
	//         }
	//       ),
	//       Service::Definition(
	//         'identifier' => TypedName(
	//           'namespace' => 'definition',
	//           'name' => 'aws_vpc::subnet'
	//         ),
	//         'serviceId' => TypedName(
	//           'namespace' => 'service',
	//           'name' => 'Yaml::Test'
	//         ),
	//         'properties' => {
	//           'parameters' => [
	//             Lyra::Parameter(
	//               'name' => 'vpcId',
	//               'type' => String
	//             ),
	//             Lyra::Parameter(
	//               'name' => 'tags',
	//               'type' => Hash[String, String]
	//             )],
	//           'returns' => [
	//             Lyra::Parameter(
	//               'name' => 'subnetId',
	//               'type' => Optional[String]
	//             )],
	//           'resourceType' => Aws::Subnet,
	//           'style' => 'resource',
	//           'origin' => '(file: testdata/aws_vpc.yaml)'
	//         }
	//       )],
	//     'style' => 'workflow',
	//     'origin' => '(file: testdata/aws_vpc.yaml)'
	//   }
	// )
	// Aws::Vpc(
	//   'amazonProvidedIpv6CidrBlock' => false,
	//   'cidrBlock' => '192.168.0.0/16',
	//   'enableDnsHostnames' => false,
	//   'enableDnsSupport' => false,
	//   'tags' => {
	//     'a' => 'av',
	//     'b' => 'bv'
	//   },
	//   'isDefault' => false,
	//   'state' => 'available'
	// )
}

func TestParse_unresolvedType(t *testing.T) {
	requireError(t, `Reference to unresolved type 'No::Such::Type' (file: testdata/typefail.yaml, line: 3, column: 5)`, func() {
		pcore.Do(func(ctx px.Context) {
			ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
			workflowFile := "testdata/typefail.yaml"
			content, err := ioutil.ReadFile(workflowFile)
			if err != nil {
				panic(err.Error())
			}
			yaml.CreateStep(ctx, workflowFile, content)
		})
	})
}

func TestParse_unparsableType(t *testing.T) {
	requireError(t, `expected one of ',' or '}', got '' (file: testdata/typeparsefail.yaml, line: 6, column: 11)`, func() {
		pcore.Do(func(ctx px.Context) {
			ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
			workflowFile := "testdata/typeparsefail.yaml"
			content, err := ioutil.ReadFile(workflowFile)
			if err != nil {
				panic(err.Error())
			}
			yaml.CreateStep(ctx, workflowFile, content)
		})
	})
}

func TestParse_mismatchedType(t *testing.T) {
	requireError(t,
		"error while building call typemismatchfail (file: testdata/typemismatchfail.yaml, line: 11, column: 7)\nCaused by: invalid arguments for function Integer: expects one of:\n  (Convertible 1, Radix 2, Boolean 3)\n    rejected: parameter 1 variant '0' expects a Numeric value, got String\n    rejected: parameter 1 variant '1' expects a Boolean value, got String\n    rejected: parameter 1 variant '2' expects a match for Pattern[/\\\\A[+-]?\\\\s*(?:(?:0|[1-9]\\\\d*)|(?:0[xX][0-9A-Fa-f]+)|(?:0[0-7]+)|(?:0[bB][01]+))\\\\z/], got 'three // Not a number'\n    rejected: parameter 1 variant '3' expects a Timespan value, got String\n    rejected: parameter 1 variant '4' expects a Timestamp value, got String\n  (NamedArgs 1)\n    rejected: parameter 1 expects a NamedArgs value, got String (file: /home/thhal/go/pkg/mod/github.com/lyraproj/pcore@v0.0.0-20190516164225-2c1838ece043/pximpl/function.go, line: 316)",
		func() {
			pcore.Do(func(ctx px.Context) {
				ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
				workflowFile := "testdata/typemismatchfail.yaml"
				content, err := ioutil.ReadFile(workflowFile)
				if err != nil {
					panic(err.Error())
				}
				yaml.CreateStep(ctx, workflowFile, content)
			})
		})
}

func TestParse_unresolvedAttr(t *testing.T) {
	requireError(t, `A Kubernetes::Namespace has no attribute named no_such_attribute (file: testdata/attrfail.yaml, line: 3, column: 14)`, func() {
		pcore.Do(func(ctx px.Context) {
			ctx.SetLoader(px.NewFileBasedLoader(ctx.Loader(), "testdata", ``, px.PuppetDataTypePath))
			workflowFile := "testdata/attrfail.yaml"
			content, err := ioutil.ReadFile(workflowFile)
			if err != nil {
				panic(err.Error())
			}
			yaml.CreateStep(ctx, workflowFile, content)
		})
	})
}

func requireError(t *testing.T, msg string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				require.Equal(t, msg, err.Error())
			} else {
				panic(r)
			}
		}
	}()
	f()
	require.Fail(t, `expected panic didn't happen`)
}
