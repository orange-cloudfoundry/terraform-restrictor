package main_test

import (
	. "github.com/orange-cloudfoundry/terraform-restrictor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"regexp"
)

var _ = Describe("Restrictor", func() {
	var rFlag RestrictorFlag
	BeforeEach(func() {
		plan := &Plan{}
		err := plan.UnmarshalFlag("./fixtures/mixed.plan")
		if err != nil {
			panic(err)
		}
		rFlag.PlanArg = PlanArg{*plan}
	})

	Context("restrict a resource by type and name", func() {
		It("should give errors about restriction when in plain text", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test1_1"),
					Type: createRegexp("credhub_generic"),
					Unauthorized: Methods{
						Method("update"),
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(1))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
		})
		It("should give errors about restriction when regex on name", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test1_[0-9]"),
					Type: createRegexp("credhub_generic"),
					Unauthorized: Methods{
						Method("update"),
					},
				},
			}

			Expect(true).To(BeTrue())
			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(2))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_2"))
		})
		It("should give errors about restriction when regex on type", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test1_1"),
					Type: createRegexp("credhub_.*"),
					Unauthorized: Methods{
						Method("update"),
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(1))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
		})
		It("should give errors about restriction when regex on type and name", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test1_[0-9]"),
					Type: createRegexp("credhub_.*"),
					Unauthorized: Methods{
						Method("update"),
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(2))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_2"))
		})
		It("should give errors about restriction when using multiple method", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test.*"),
					Type: createRegexp("credhub_(generic|password)"),
					Unauthorized: Methods{
						Method("update"),
						Method("create"),
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(3))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("unauthorized to use method 'create'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_2"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_password.test3"))
		})
	})
	Context("restrict attribute by method", func() {
		It("should give errors about restriction on 1 attribute when given", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test1_1"),
					Type: createRegexp("credhub_generic"),
					CheckAttrs: CheckAttrs{
						{
							Path: createRegexp("data_json"),
							Unauthorized: Methods{
								Method("update"),
							},
						},
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(1))
			Expect(merr.Error()).Should(ContainSubstring("'data_json' is unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
		})
		It("should give errors about restriction on multiple attributes when given", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test.*"),
					Type: createRegexp("credhub_.*"),
					CheckAttrs: CheckAttrs{
						{
							Path: createRegexp("data_json"),
							Unauthorized: Methods{
								Method("update"),
							},
						},
						{
							Path: createRegexp("name"),
							Unauthorized: Methods{
								Method("create"),
							},
						},
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(3))
			Expect(merr.Error()).Should(ContainSubstring("'data_json' is unauthorized to use method 'update'"))
			Expect(merr.Error()).Should(ContainSubstring("'name' is unauthorized to use method 'create'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_generic.test1_1"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_password.test3"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_rsa.test4"))
		})
	})
	Context("validating attribute against regex", func() {
		It("should give errors about validation with 1 validation", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test.*"),
					Type: createRegexp("credhub_.*"),
					CheckAttrs: CheckAttrs{
						{
							Path:     createRegexp("name"),
							Validate: createRegexps(".*test-failed.*"),
						},
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(2))
			Expect(merr.Error()).Should(ContainSubstring("it should match regex '/^.*test-failed.*$/' and it was '/test/3'"))
			Expect(merr.Error()).Should(ContainSubstring("it should match regex '/^.*test-failed.*$/' and it was '/test/4'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_password.test3"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_rsa.test4"))
		})
		It("should give errors about validation with mulitple validation", func() {
			rFlag.Restrictions = Resources{
				{
					Name: createRegexp("test.*"),
					Type: createRegexp("credhub_.*"),
					CheckAttrs: CheckAttrs{
						{
							Path:     createRegexp("name"),
							Validate: createRegexps(".*test-failed.*", ".*test-failed2.*"),
						},
					},
				},
			}

			err := CheckRestrictions(rFlag)
			Expect(err).To(HaveOccurred())

			merr := err.(MultipleErrors)
			Expect(merr).Should(HaveLen(2))
			Expect(merr.Error()).Should(ContainSubstring("it should match regex '/^(.*test-failed.*)|(.*test-failed2.*)$/' and it was '/test/3'"))
			Expect(merr.Error()).Should(ContainSubstring("it should match regex '/^(.*test-failed.*)|(.*test-failed2.*)$/' and it was '/test/4'"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_password.test3"))
			Expect(merr.Error()).Should(ContainSubstring("credhub_rsa.test4"))
		})
	})
})

func createRegexp(s string) Regexp {
	regex := regexp.MustCompile("^(?:" + s + ")$")
	re := Regexp{}
	re.Raw = s
	re.Regexp = regex
	return re
}

func createRegexps(s ...string) Regexps {
	res := make(Regexps, 0)
	for _, el := range s {
		res = append(res, createRegexp(el))
	}
	return res
}
