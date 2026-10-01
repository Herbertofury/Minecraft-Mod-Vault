package net.fabric_extras.ranged_weapon.internal;

/** Public 2.3.x haste bridge retained for downstream compatibility. */
public interface RangedHasteEntity {
    void resetPartialHasteTicks();
    float getPartialHasteTick();
    void addPartialHasteTick(float tick);
}
