#version 330 core

out vec4 FragColor;

in vec3 FragPos;
in vec3 Color;

// Фонарик (SpotLight от игрока)
uniform bool flashlightEnabled;
uniform vec3 flashlightPos;
uniform vec3 flashlightDir;
uniform vec3 flashlightColor;

// Точечный свет в центре
uniform bool centerLightEnabled;
uniform vec3 centerLightPos;
uniform vec3 centerLightColor;

// Ambient
uniform vec3 ambientColor;
uniform float ambientStrength;

void main()
{
    // Ambient lighting
    vec3 ambient = ambientColor * ambientStrength;

    vec3 lighting = ambient;

    // Фонарик (SpotLight)
    if (flashlightEnabled) {
        vec3 lightDir = normalize(flashlightPos - FragPos);
        float distance = length(flashlightPos - FragPos);

        // Attenuation
        float attenuation = 1.0 / (1.0 + 0.09 * distance + 0.032 * (distance * distance));

        // Spotlight cone
        float theta = dot(lightDir, normalize(-flashlightDir));
        float cutOff = cos(radians(12.5)); // Внутренний угол
        float outerCutOff = cos(radians(17.5)); // Внешний угол
        float epsilon = cutOff - outerCutOff;
        float intensity = clamp((theta - outerCutOff) / epsilon, 0.0, 1.0);

        // Простое диффузное освещение без нормалей - используем расстояние
        float diffuse = max(1.0 - distance / 20.0, 0.0);

        vec3 flashlight = flashlightColor * diffuse * attenuation * intensity * 2.0;
        lighting += flashlight;
    }

    // Точечный свет в центре
    if (centerLightEnabled) {
        vec3 lightDir = normalize(centerLightPos - FragPos);
        float distance = length(centerLightPos - FragPos);

        // Attenuation
        float attenuation = 1.0 / (1.0 + 0.09 * distance + 0.032 * (distance * distance));

        // Простое диффузное освещение
        float diffuse = max(1.0 - distance / 25.0, 0.0);

        vec3 pointLight = centerLightColor * diffuse * attenuation * 3.0;
        lighting += pointLight;
    }

    // Итоговый цвет
    vec3 result = Color * lighting;

    FragColor = vec4(result, 1.0);
}
