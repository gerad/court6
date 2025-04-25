import numpy as np
import trimesh
import math
import os

# Constants (in meters)
COURT_LENGTH = 23.77  # 78ft
COURT_WIDTH_SINGLES = 8.23  # 27ft
COURT_WIDTH_DOUBLES = 10.97  # 36ft
SERVICE_LINE_DISTANCE = 6.4  # 21ft
NET_HEIGHT_POSTS = 1.067  # 42in
NET_HEIGHT_CENTER = 0.914  # 36in
NET_WIDTH = 12.8  # 42ft
NET_DEPTH = 0.003  # ~3mm
NET_POST_DIAMETER = 0.07  # ~2.75in
LINE_WIDTH = 0.051  # 2in
BASELINE_WIDTH = 0.102  # 4in

def create_line_segment(x1, y1, z1, x2, y2, z2, width):
    """Create a line segment as a thin box"""
    # Calculate direction vector
    dx = x2 - x1
    dy = y2 - y1
    dz = z2 - z1
    length = math.sqrt(dx*dx + dy*dy + dz*dz)
    
    # Normalize
    dx /= length
    dy /= length
    dz /= length
    
    # Create perpendicular vectors for the width
    # We'll use the cross product with the up vector (0,1,0) to get a perpendicular vector
    px, py, pz = np.cross([dx, dy, dz], [0, 1, 0])
    # Normalize
    length_perp = math.sqrt(px*px + py*py + pz*pz)
    px /= length_perp
    py /= length_perp
    pz /= length_perp
    
    # Scale by half width
    half_width = width / 2
    px *= half_width
    py *= half_width
    pz *= half_width
    
    # Create the 8 vertices of the box
    vertices = np.array([
        [x1 + px, y1 + py, z1 + pz],  # Top left
        [x1 - px, y1 - py, z1 - pz],  # Bottom left
        [x2 + px, y2 + py, z2 + pz],  # Top right
        [x2 - px, y2 - py, z2 - pz],  # Bottom right
        [x1 + px, y1 + py + width, z1 + pz],  # Top left (raised)
        [x1 - px, y1 - py + width, z1 - pz],  # Bottom left (raised)
        [x2 + px, y2 + py + width, z2 + pz],  # Top right (raised)
        [x2 - px, y2 - py + width, z2 - pz],  # Bottom right (raised)
    ])
    
    # Define the faces (triangles)
    faces = np.array([
        [0, 1, 2], [1, 3, 2],  # Bottom face
        [4, 5, 6], [5, 7, 6],  # Top face
        [0, 4, 2], [2, 4, 6],  # Front face
        [1, 3, 5], [3, 7, 5],  # Back face
        [0, 1, 4], [1, 5, 4],  # Left face
        [2, 6, 3], [3, 6, 7],  # Right face
    ])
    
    return trimesh.Trimesh(vertices=vertices, faces=faces)

def create_cylinder(x, y, z, radius, height):
    """Create a cylinder"""
    # Create a cylinder using trimesh's built-in primitive
    cylinder = trimesh.creation.cylinder(radius=radius, height=height)
    
    # Move the cylinder to the desired position
    cylinder.apply_translation([x, y, z])
    
    return cylinder

def create_tennis_court():
    # Create an empty scene
    scene = trimesh.Scene()
    
    # Draw the court outline (doubles court)
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, 0, -COURT_LENGTH/2, 
        COURT_WIDTH_DOUBLES/2, 0, -COURT_LENGTH/2, 
        BASELINE_WIDTH
    ))  # Baseline
    
    scene.add_geometry(create_line_segment(
        COURT_WIDTH_DOUBLES/2, 0, -COURT_LENGTH/2, 
        COURT_WIDTH_DOUBLES/2, 0, COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Right sideline
    
    scene.add_geometry(create_line_segment(
        COURT_WIDTH_DOUBLES/2, 0, COURT_LENGTH/2, 
        -COURT_WIDTH_DOUBLES/2, 0, COURT_LENGTH/2, 
        BASELINE_WIDTH
    ))  # Baseline
    
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, 0, COURT_LENGTH/2, 
        -COURT_WIDTH_DOUBLES/2, 0, -COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Left sideline
    
    # Draw singles court lines
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_SINGLES/2, 0, -COURT_LENGTH/2, 
        COURT_WIDTH_SINGLES/2, 0, -COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Singles baseline
    
    scene.add_geometry(create_line_segment(
        COURT_WIDTH_SINGLES/2, 0, -COURT_LENGTH/2, 
        COURT_WIDTH_SINGLES/2, 0, COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Singles right sideline
    
    scene.add_geometry(create_line_segment(
        COURT_WIDTH_SINGLES/2, 0, COURT_LENGTH/2, 
        -COURT_WIDTH_SINGLES/2, 0, COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Singles baseline
    
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_SINGLES/2, 0, COURT_LENGTH/2, 
        -COURT_WIDTH_SINGLES/2, 0, -COURT_LENGTH/2, 
        LINE_WIDTH
    ))  # Singles left sideline
    
    # Draw service lines
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, 0, -SERVICE_LINE_DISTANCE, 
        COURT_WIDTH_DOUBLES/2, 0, -SERVICE_LINE_DISTANCE, 
        LINE_WIDTH
    ))  # Service line
    
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, 0, SERVICE_LINE_DISTANCE, 
        COURT_WIDTH_DOUBLES/2, 0, SERVICE_LINE_DISTANCE, 
        LINE_WIDTH
    ))  # Service line
    
    # Draw center line
    scene.add_geometry(create_line_segment(
        0, 0, -SERVICE_LINE_DISTANCE, 
        0, 0, SERVICE_LINE_DISTANCE, 
        LINE_WIDTH
    ))  # Center line
    
    # Draw center mark
    scene.add_geometry(create_line_segment(
        -0.1, 0, 0, 
        0.1, 0, 0, 
        LINE_WIDTH
    ))  # Center mark
    
    # Draw net posts
    post_radius = NET_POST_DIAMETER / 2
    post_height = NET_HEIGHT_POSTS
    
    # Left net post
    scene.add_geometry(create_cylinder(
        -COURT_WIDTH_DOUBLES/2, 0, 0, 
        post_radius, post_height
    ))
    
    # Right net post
    scene.add_geometry(create_cylinder(
        COURT_WIDTH_DOUBLES/2, 0, 0, 
        post_radius, post_height
    ))
    
    # Net top
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, NET_HEIGHT_POSTS, 0, 
        COURT_WIDTH_DOUBLES/2, NET_HEIGHT_POSTS, 0, 
        NET_DEPTH
    ))
    
    # Net center tape
    scene.add_geometry(create_line_segment(
        -COURT_WIDTH_DOUBLES/2, 0, 0, 
        COURT_WIDTH_DOUBLES/2, 0, 0, 
        NET_DEPTH
    ))
    
    # Net with center sag
    # We'll approximate the net with a few segments
    segments = 10
    for i in range(segments + 1):
        x = -COURT_WIDTH_DOUBLES/2 + (COURT_WIDTH_DOUBLES / segments) * i
        # Calculate height with parabolic sag
        t = i / segments
        height = NET_HEIGHT_POSTS - (NET_HEIGHT_POSTS - NET_HEIGHT_CENTER) * 4 * t * (1 - t)
        
        if i > 0:
            prev_x = -COURT_WIDTH_DOUBLES/2 + (COURT_WIDTH_DOUBLES / segments) * (i - 1)
            prev_height = NET_HEIGHT_POSTS - (NET_HEIGHT_POSTS - NET_HEIGHT_CENTER) * 4 * (i - 1) / segments * (1 - (i - 1) / segments)
            scene.add_geometry(create_line_segment(
                prev_x, prev_height, 0, 
                x, height, 0, 
                NET_DEPTH
            ))
    
    return scene

def export_glb(scene, filename="tennis_court.glb"):
    """Export the scene as a GLB file"""
    # Check if we're running in Docker
    if os.path.exists("/app/output"):
        output_path = os.path.join("/app/output", filename)
    else:
        output_path = filename
        
    scene.export(output_path, file_type="glb")
    print(f"Tennis court model exported to {output_path}")

if __name__ == "__main__":
    # Create the tennis court scene
    court_scene = create_tennis_court()
    
    # Export as GLB
    export_glb(court_scene)
    
    print("Tennis court GLB model created successfully!")
