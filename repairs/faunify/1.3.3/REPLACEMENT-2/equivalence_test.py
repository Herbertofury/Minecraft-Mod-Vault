import random

R = random.Random(1337)

def original(cands):
    # Stream.min(comparingInt(-light).thenComparingDouble(distance))
    best = None
    best_key = None
    for c in cands:
        if c['blacklisted'] or c['emission'] < 8:
            continue
        key = (-c['light'], c['distance'])
        if best is None or key < best_key:
            best = c['id']
            best_key = key
    return best

def patched(cands):
    best = None
    best_light = -1
    best_distance = float('inf')
    for c in cands:
        if c['blacklisted'] or c['emission'] < 8:
            continue
        light = c['light']
        distance = c['distance']
        if light > best_light or (light == best_light and distance < best_distance):
            best = c['id']
            best_light = light
            best_distance = distance
    return best

for case in range(10000):
    n = R.randint(0, 512)
    cands=[]
    for i in range(n):
        cands.append({
            'id': i,
            'blacklisted': R.random() < 0.17,
            'emission': R.randint(0, 15),
            'light': R.randint(0, 15),
            'distance': R.choice([float(R.randint(0, 900)), R.random()*900.0]),
        })
    a=original(cands); b=patched(cands)
    if a != b:
        raise SystemExit(f'FAIL case={case} original={a} patched={b}')
print('PASS: 10000 randomized selection/tie/threshold/blacklist cases matched exactly')
