# Pod 基础

> 详细请见：[K8s-Pod](https://hedon954.github.io/noteSite/linux/k8s/k8s-pod.html)



## Pod 状态

| 取值                | 描述                                                         |
| :------------------ | :----------------------------------------------------------- |
| `Pending`（悬决）   | Pod 已被 Kubernetes 系统接受，但有一个或者多个容器尚未创建亦未运行。此阶段包括等待 Pod 被调度的时间和通过网络下载镜像的时间。 |
| `Running`（运行中） | Pod 已经绑定到了某个节点，Pod 中所有的容器都已被创建。至少有一个容器仍在运行，或者正处于启动或重启状态。 |
| `Succeeded`（成功） | Pod 中的所有容器都已成功终止，并且不会再重启。               |
| `Failed`（失败）    | Pod 中的所有容器都已终止，并且至少有一个容器是因为失败终止。也就是说，容器以非 0 状态退出或者被系统终止。 |
| `Unknown`（未知）   | 因为某些原因无法取得 Pod 的状态。这种情况通常是因为与 Pod 所在主机通信失败。 |

> 如果某节点死掉或者与集群中其他节点失联，Kubernetes 会实施一种策略，将失去的节点上运行的所有 Pod 的 `phase` 设置为 `Failed`。



## Pod 镜像拉取策略

- Always：总是从远程仓库拉取镜像
- IfNotPresent：本地有则使用本地镜像，本地没有则从远程仓库拉取镜像
- Never：只使用本地镜像，从不去远程仓库拉取，本地没有就报错

> 默认值说明：
>
> - 如果镜像 tag 为具体的版本号，默认策略是 IfNotPresent
> - 如果镜像 tag 为 latest（最终版本），默认策略是 Always



## Pod 重启策略

Pod 的 `spec` 中包含一个 `restartPolicy` 字段，其可能取值包括：

- **Always**（默认）
- **OnFailure**
- **Never**

`restartPolicy` 适用于 Pod 中的所有容器。`restartPolicy` 仅针对同一节点上 `kubelet`的容器重启动作。

首次需要重启的容器，将在其需要的时候立即进行重启，随后再次重启的操作将由 kubelet 延迟一段时间后进行，且反复的重启操作的延迟时长以此为 10s、20s、40s、80s、160s 和 300s，300s 是最大的延迟时长。

一旦某容器执行了 10 分钟并且没有出现问题，`kubelet` 对该容器的重启回退计时器执行重置操作。

> - RC 和 DeamonSet 必须设置为 Always，需要保证该容器持续运行；
> - Job：OnFailure 或 Never，确保容器执行完成后不再重启；



## Pod 发布策略

- `recreate 重建`：停止旧版本，部署新版本
- `rolling-update 滚动更新`：一个接一个地滚动更新方式发布新版本
- `blue/green 蓝绿`：新版本与旧版本一起存在，然后切换流量
- `canary 金丝雀`：将新版本向一部分用户发布，然后继续全量发布
- `a/b testing A/B  测`：以精确的方式（HTTP Header、cookie、权重等）向部分用户发布新版本。A/B 测实际上是一种基于数据统计做出业务决策的技术。在 k8s 中并不原生支持，需要额外的一些高级组件来完成该设置。比如 Istio、Linkerd、Traefik，或者自定义 Nginx/Haproxy 等。

