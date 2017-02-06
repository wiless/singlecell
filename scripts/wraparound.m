directions = [[1, -1, 0]; [1, 0, -1]; [0, +1, -1]; [-1, +1, 0]; [-1, 0, +1]; [0, -1, +1]]
FINALRADIUS=2;
mirrorCenter= [2*FINALRADIUS+1, -FINALRADIUS, -FINALRADIUS-1]

mirrorTables(1,:)=mirrorCenter;
oldcenter=mirrorCenter;
for k=2:6    
     
    newcenter=-[oldcenter(2:end) oldcenter(1) ];
    mirrorTables(k,:)=newcenter;
    oldcenter=newcenter;
end


mirrorTables

