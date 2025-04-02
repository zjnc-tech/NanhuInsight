<h1 align="center" style="border-bottom: none">
    <a href="" target="_blank">
        <img alt="NanhuInsight" src="" width="100" height="100">
    </a>
    <br>NanhuInsight
</h1>

<p align="center">NanhuInsight是一个专为大规模智算集群设计的智能故障控制与高效运维的开源项目。它提供从遥测信号采集、性能监控、故障告警、健康检查到智能故障分析与自动处置的一站式解决方案，帮助降低AI训练任务中的故障影响，提升集群运行稳定性与运维效率</p>
  
<h3 align="center">
  <a href="http://pre-isg.doc.zhejianglab.org/"><b>Documentation</b></a> &bull;
  <a href="https://github.com/SigNoz/signoz/blob/main/README.en.md"><b>ReadMe in English</b></a> &bull;
</h3>

## 核心特性

### 📊 遥测数据采集

基于vector实现指标与日志的OneAgent式的高性能遥测数据管道；收集、转换和路由所有的日志、指标数据；采用混合精度采集策略，区分故障判别的核心指标与常规指标；它提供针对智算集群优化过的开箱即用的采集器，同时也兼容opentelemetry、victoria metrics servicescrape等采集协议以支持扩展。

### 🖥️ 集群监控

提供一屏化监控、下钻式拓扑可视化、历史数据回溯，支持指标、日志的检索与可视化分析，提升运维人员运维效率。

### 🚨 故障告警

当集群中发生任何异常时，会收到通知；可以针对任何类型的遥测信号（日志、指标）设置警报，创建阈值并设置通知渠道以获取通知。系统将自动对告警的严重等级、影响范围进行评估，对节点健康度进行打分；系统在k8s控制平面以crd的形式暴露集群故障诊断结果，作业平台、运维人员可基于此对故障进行半自动化的处置。

### 🩺 主动健康检查

使用日志、指标等被动式的告警难以将故障覆盖完全，且一些隐晦的故障无法通过被动式的告警检出。所以NanhuInsight项目还提供日常巡检、深度检查、训前压测等多种主动式健康检查方式，确保作业运行前无潜在隐患。

### 🤖 运维智能体

基于LLM的多智能体将会在系统的各个环节发挥效果，提供运维知识库、智能故障判定、根因分析、自动化处置，增强系统自治能力。

### 🚀 多集群统一监控

一个数据中心有时会存在多个异构的GPU集群，NanhuInsight提供联邦化的故障控制能力，并且已经适配了NVIDIA、摩尔、沐曦、昆仑芯等厂商，可以在同一个页面上获得一致的上述功能的体验，避免与厂商提供的监控系统绑定而需要学习多个监控系统。

## 软件架构

南湖可观测在管控集群和计算集群都会部署组件。

计算集群的每个计算节点中都会运行一个Vector Agent，负责采集机器硬件的指标、容器和操作系统日志。还会一个健康检查agent，负责执行健康检查任务。

计算集群的边缘会运行一个Vector Aggregator，负责采集平台的指标并聚合Vector Agent的数据写入位于管控集群的存储。还会运行一个健康检查Executor，负责下发健康检查任务。

管控集群中会运行Clickhouse cluster负责数据存储。

管控集群中还会运行一个ODRE负责对数据进行关联处理、故障告警，提供数据接口。

管控集群中还会运行一个主动健康检查调度器，负责健康检查任务的调度。

管控集群中还会运行一个运维智能体，负责自动化的故障定位和处置。

<p align="center">
  <img src="docs/overview/_static/nanhu软件架构.png" alt="NanhuInsight Architecture" width="500">
</p>
