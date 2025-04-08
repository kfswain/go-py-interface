import math


def selectPod(pods):
    ret_list = []
    for pod in pods:
        score = 0
        score += int(math.log(pod["kv_cache_util"])) * 10
        score += pod["queue_count"] * -10_000

        ret_list.append((pod["pod_name"], score))
        
    ret_list = sorted(ret_list, key=lambda tup: tup[1], reverse=True)

    return ret_list