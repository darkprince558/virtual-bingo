import { useState } from "react";
import { generateTheme } from "../api/themeApi";


export function useThemeGenerator() {

    const [theme, setTheme] = useState();

    const createTheme = async (
        topic: string
    ) => {

        const result =
            await generateTheme(topic);

        setTheme(result);
    };

    return {
        theme,
        createTheme
    };
}
