<script lang="ts">
    export let score: number;

    let color = "bg-slate-100 text-slate-600";
    let label = "Low";

    $: {
        if (score >= 0.85) {
            color = "bg-emerald-100 text-emerald-700 border-emerald-200";
            label = "High";
        } else if (score >= 0.6) {
            color = "bg-amber-100 text-amber-700 border-amber-200";
            label = "Medium";
        } else {
            color = "bg-slate-100 text-slate-600 border-slate-200";
            label = "Low";
        }
    }
</script>

<span
    class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border {color}"
>
    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 10V3L4 14h7v7l9-11h-7z"
        />
    </svg>
    {Math.round(score * 100)}%
</span>
