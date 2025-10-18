#version 330 core

out vec4 FragColor;

in vec3 FragPos;
in vec3 Normal;
in vec3 Color;
in vec4 FragPosLightSpace;

// Максимум 4 источника света для оптимизации
#define MAX_LIGHTS 4

// Типы света
#define LIGHT_DIRECTIONAL 0
#define LIGHT_POINT 1
#define LIGHT_SPOT 2

struct Light {
    int type;
    vec3 position;
    vec3 direction;
    vec3 color;
    float intensity;

    // Для SpotLight
    float cutOff;
    float outerCutOff;

    // Для PointLight
    float constant;
    float linear;
    float quadratic;
};

uniform vec3 viewPos;
uniform vec3 ambientColor;
uniform float ambientStrength;

uniform int numLights;
uniform Light lights[MAX_LIGHTS];

uniform sampler2D shadowMap;
uniform bool useShadows;

// Оптимизированный расчёт теней с PCF (Percentage Closer Filtering)
float ShadowCalculation(vec4 fragPosLightSpace)
{
    if (!useShadows)
        return 0.0;

    // Perspective divide
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;

    // Transform to [0,1] range
    projCoords = projCoords * 0.5 + 0.5;

    // Вне зоны shadow map - нет тени
    if(projCoords.z > 1.0)
        return 0.0;

    // Текущая глубина фрагмента
    float currentDepth = projCoords.z;

    // Bias для устранения shadow acne
    float bias = 0.005;

    // PCF (Percentage Closer Filtering) для мягких теней
    // Используем 2x2 grid для оптимизации (можно увеличить до 3x3 или 5x5)
    float shadow = 0.0;
    vec2 texelSize = 1.0 / textureSize(shadowMap, 0);

    for(int x = -1; x <= 1; ++x)
    {
        for(int y = -1; y <= 1; ++y)
        {
            float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r;
            shadow += currentDepth - bias > pcfDepth ? 1.0 : 0.0;
        }
    }
    shadow /= 9.0; // Усредняем по 9 сэмплам

    return shadow;
}

// Оптимизированный расчёт направленного света
vec3 CalcDirectionalLight(Light light, vec3 normal, vec3 viewDir, float shadow)
{
    vec3 lightDir = normalize(-light.direction);

    // Diffuse (диффузное освещение) - модель Ламберта
    float diff = max(dot(normal, lightDir), 0.0);

    // Specular (зеркальное отражение) - модель Блинна-Фонга (оптимизированная)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 32.0);

    // Комбинируем
    vec3 diffuse = light.color * diff * light.intensity;
    vec3 specular = light.color * spec * light.intensity * 0.3; // Умножаем на 0.3 для меньшего блеска

    // Применяем тени
    return (1.0 - shadow) * (diffuse + specular);
}

// Оптимизированный расчёт точечного света
vec3 CalcPointLight(Light light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);

    // Diffuse
    float diff = max(dot(normal, lightDir), 0.0);

    // Specular (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 32.0);

    // Attenuation (затухание)
    float distance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));

    // Комбинируем
    vec3 diffuse = light.color * diff * light.intensity;
    vec3 specular = light.color * spec * light.intensity * 0.3;

    return attenuation * (diffuse + specular);
}

// Оптимизированный расчёт прожектора
vec3 CalcSpotLight(Light light, vec3 normal, vec3 fragPos, vec3 viewDir)
{
    vec3 lightDir = normalize(light.position - fragPos);

    // Diffuse
    float diff = max(dot(normal, lightDir), 0.0);

    // Specular
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 32.0);

    // Attenuation
    float distance = length(light.position - fragPos);
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));

    // Spotlight intensity (мягкие края)
    float theta = dot(lightDir, normalize(-light.direction));
    float epsilon = light.cutOff - light.outerCutOff;
    float intensity = clamp((theta - light.outerCutOff) / epsilon, 0.0, 1.0);

    // Комбинируем
    vec3 diffuse = light.color * diff * light.intensity;
    vec3 specular = light.color * spec * light.intensity * 0.3;

    return attenuation * intensity * (diffuse + specular);
}

void main()
{
    vec3 norm = normalize(Normal);
    vec3 viewDir = normalize(viewPos - FragPos);

    // Ambient lighting (окружающее освещение)
    vec3 ambient = ambientColor * ambientStrength;

    // Расчёт теней (только для первого источника света)
    float shadow = ShadowCalculation(FragPosLightSpace);

    // Расчёт освещения от всех источников
    vec3 lighting = vec3(0.0);

    for(int i = 0; i < numLights && i < MAX_LIGHTS; i++)
    {
        if(lights[i].type == LIGHT_DIRECTIONAL)
        {
            // Тени только для первого направленного света
            float lightShadow = (i == 0) ? shadow : 0.0;
            lighting += CalcDirectionalLight(lights[i], norm, viewDir, lightShadow);
        }
        else if(lights[i].type == LIGHT_POINT)
        {
            lighting += CalcPointLight(lights[i], norm, FragPos, viewDir);
        }
        else if(lights[i].type == LIGHT_SPOT)
        {
            lighting += CalcSpotLight(lights[i], norm, FragPos, viewDir);
        }
    }

    // Итоговый цвет = базовый цвет * (ambient + освещение)
    vec3 result = Color * (ambient + lighting);

    FragColor = vec4(result, 1.0);
}
