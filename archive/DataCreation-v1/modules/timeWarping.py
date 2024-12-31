def timeWarp(time, referenceTime, factor) -> int:
    duration = time - referenceTime
    duration /= factor
    time = int(referenceTime + duration)
    return time
