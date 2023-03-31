<template>
  <div :style="{ display: 'flex' }">
    <a-card :style="{ width: '360px' }" title="Arco Card">

      <form-create :rule="rule" v-model:api="fApi" :option="option"></form-create>
      formats.
    </a-card>
  </div>

</template>

<script>
import { reactive, ref, onMounted } from "vue";
export default {
  setup() {
    onMounted(() => {
      for (const key in operatorData) {
        if (Object.prototype.hasOwnProperty.call(operatorData, key)) {
          operatorType.push(key);
        }
      }
    });
    let operatorData = reactive({
      ipaddr: [
        {
          key_name: "nicName",
          description: "vip binding nic",
          validation_tags: "required | string",
        },
        {
          key_name: "ipaddr",
          description: "vip ipaddr",
          validation_tags: "required | isIP",
        },
        {
          key_name: "netmask",
          description: "vip netmask",
          validation_tags: "required | isFloat",
        },
        {
          key_name: "gateway",
          description: "vip gateway",
          validation_tags: "required | isIP",
        },
      ],
      mock: null,
      nfs: [
        {
          key_name: "sharedInfoDir",
          description: "service name",
          validation_tags: "required | string",
        },
      ],
      process: [
        {
          key_name: "RunCmd",
          description: "The full path of command to run",
          validation_tags: "required",
        },
      ],
      service: [
        {
          key_name: "serviceName",
          description: "service name",
          validation_tags: "required | string",
        },
      ],
      svo: [
        {
          key_name: "test1",
          description: "test1",
          validation_tags: "required",
        },
        {
          key_name: "test2",
          description: "test2",
          validation_tags:
              "required_if:test1,testabc | required_without:foo,bar",
        },
        {
          key_name: "mail",
          description: "The email address",
          validation_tags: "required|email",
        },
        {
          key_name: "enable",
          description: "Enable or not",
          validation_tags: "required | bool",
        },
        {
          key_name: "number",
          description: "The number",
          validation_tags: "required|number| between:249,251",
        },
      ],
      zfs: [
        {
          key_name: "poolName",
          description: "zfs pool name",
          validation_tags: "required | string",
        },
        {
          key_name: "checkTime",
          description: "zfs export times",
          validation_tags: "required | isFloat",
        },
      ],
    });
    const operatorType = reactive([]);
    const operatorRules = ref([]);
    const rule = ref([
      {
        type: "input",
        field: "id",
        title: "ID",
        value: "",
        effect: {
          required: "请输入ID",
        },
        validate: [{ type: "string", required: true, message: "请输入ID" }],
      },
      {
        type: "select",
        field: "kind",
        title: "类型",
        value: "",
        effect: {
          required: "请选择类型",
        },
        validate: [{ type: "string", required: true, message: "请选择类型" }],
        options: operatorType,
        update(val, rule, fApi) {
          operatorTypeChange(val);
        },
        control: [
          {
            handle: (val) => !!val,
            append: "kind",
            rule: [],
          },
        ],
      },
    ]);
    const fApi = ref({});
    const option = ref({
      submitBtn: {
        show: true  ,
      },
      form: {
        layout: "horizontal",
        labelAlign: "right",
        labelColProps: {
          span: 4,
        },
        wrapperColProps: {
          span: 20,
        },
      },
    });
    const operatorTypeChange = (e) => {
      rule.value[1].control[0].rule = [];
      console.log("control-----", rule.value[1].control[0].rule);
      rule.value[1].control[0].rule = operatorRules.value = [];
      const element = operatorData[e];
      element.forEach((item) => {
        let obj = {
          key: item.key_name,
          type: "input",
          field: item.key_name,
          title: item.key_name,
          value: "",
          validate: [],
        };
        operatorRules.value.push(obj);
      });
      console.log("-----", operatorRules.value);
      rule.value[1].control[0].rule = operatorRules.value;
      fApi.value.refresh();
      fApi.value.nextRefresh(() => {
        //todo 表单操作
      });
    };
    return {
      rule,
      option,
      fApi,
    };
  },
};
</script>

<style>
</style>
