#version 410 core
in vec4 Col;
in vec3 Nor;
in vec2 TexCoords;
flat in int ShapeType;
void main() {
	float dotprod = dot(TexCoords, TexCoords);
	gl_FragColor = vec4(dotprod, 2 - dotprod, dotprod*dotprod, 1.0);
}
