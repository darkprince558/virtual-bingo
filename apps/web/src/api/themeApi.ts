export async function generateTheme(
    topic: string
) {

    const response = await fetch(
        `/api/theme/generate-theme/${encodeURIComponent(topic)}`
    );

    if (!response.ok) {
        throw new Error(
            "Unable to generate theme"
        );
    }

    return response.json();
}
