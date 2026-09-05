package net.fabric_extras.ranged_weapon.internal;

/** Tracks whether Ranged Weapon API has already applied projectile damage scaling. */
public interface ArrowExtension {
    void rwa_markModified(boolean modified);
    boolean rwa_isModified();
}
