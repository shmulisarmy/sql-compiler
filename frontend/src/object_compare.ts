// Deep structural equality check for objects, arrays, primitives
export function deepObjectCompare(a: any, b: any): boolean {
    if (a === b) return true;

    // Handle NaN
    if (typeof a === "number" && typeof b === "number") {
        return Number.isNaN(a) && Number.isNaN(b);
    }

    if (typeof a !== typeof b) return false;
    if (a === null || b === null) return false;

    // Arrays
    if (Array.isArray(a) && Array.isArray(b)) {
        if (a.length !== b.length) return false;
        for (let i = 0; i < a.length; i++) {
            if (!deepObjectCompare(a[i], b[i])) return false;
        }
        return true;
    }

    // Objects
    if (typeof a === "object" && typeof b === "object") {
        const aKeys = Object.keys(a);
        const bKeys = Object.keys(b);
        if (aKeys.length !== bKeys.length) return false;

        for (const key of aKeys) {
            if (!b.hasOwnProperty(key)) return false;
            if (!deepObjectCompare(a[key], b[key])) return false;
        }

        return true;
    }

    return false;
}
